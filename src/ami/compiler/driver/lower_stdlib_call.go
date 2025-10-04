package driver

import (
    "strings"
    "strconv"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerStdlibCall recognizes AMI stdlib calls and lowers them.
func lowerStdlibCall(st *lowerState, c *ast.CallExpr) (ir.Expr, bool) {
    if c == nil { return ir.Expr{}, false }
    name := c.Name
    if ex, ok := lowerStdlibMath(st, c); ok { return ex, true }
    if (name == "signal.Register") || (st != nil && st.funcParams != nil && len(st.funcParams[name]) == 2 && st.funcParams[name][0] == "SignalType") {
        if len(c.Args) >= 2 {
            var args []ir.Value
            // arg0: signal type (enum) → prefer immediate OS mapping when selector provided
            switch s := c.Args[0].(type) {
            case *ast.SelectorExpr:
                var v int64
                switch s.Sel { case "SIGINT": v = 2; case "SIGTERM": v = 15; case "SIGHUP": v = 1; case "SIGQUIT": v = 3 }
                if v != 0 {
                    args = append(args, ir.Value{ID: "#"+strconv.FormatInt(v, 10), Type: "int64"})
                }
            }
            if len(args) == 0 {
                if ex, ok := lowerExpr(st, c.Args[0]); ok && ex.Result != nil { args = append(args, ir.Value{ID: ex.Result.ID, Type: "int64"}) }
            }
            // arg1: handler token (deterministic immediate)
            tok, ok := handlerTokenValue(c.Args[1])
            if !ok { tok = 0 }
            args = append(args, ir.Value{ID: "#"+strconv.FormatInt(tok, 10), Type: "int64"})
            return ir.Expr{Op: "call", Callee: "ami_rt_signal_register", Args: args}, true
        }
        return ir.Expr{Op: "call", Callee: "ami_rt_signal_register", Args: []ir.Value{{ID: "#0", Type: "int64"}, {ID: "#0", Type: "int64"}}}, true
    }
    if strings.HasSuffix(name, ".MetalBlockingSubmit") || name == "gpu.MetalBlockingSubmit" {
        if len(c.Args) == 1 {
            if inner, ok := c.Args[0].(*ast.CallExpr); ok {
                if ex, ok2 := lowerExpr(st, inner); ok2 { return ex, true }
            }
        }
        var args []ir.Value
        for _, a := range c.Args { if ex, ok := lowerExpr(st, a); ok && ex.Result != nil { args = append(args, *ex.Result) } }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "Error<any>"}
        return ir.Expr{Op: "call", Callee: "ami_rt_gpu_blocking_submit", Args: args, Result: res}, true
    }
    if strings.HasSuffix(name, ".MetalAvailable") || name == "gpu.MetalAvailable" {
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "bool"}
        // use mask accessor bit 0
        return ir.Expr{Op: "call", Callee: "ami_rt_gpu_has", Args: []ir.Value{{ID: "#0", Type: "int64"}}, Result: res}, true
    }
    if strings.HasSuffix(name, ".CudaAvailable") || name == "gpu.CudaAvailable" {
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "bool"}
        // use mask accessor bit 1
        return ir.Expr{Op: "call", Callee: "ami_rt_gpu_has", Args: []ir.Value{{ID: "#1", Type: "int64"}}, Result: res}, true
    }
    if strings.HasSuffix(name, ".OpenCLAvailable") || name == "gpu.OpenCLAvailable" {
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "bool"}
        // use mask accessor bit 2
        return ir.Expr{Op: "call", Callee: "ami_rt_gpu_has", Args: []ir.Value{{ID: "#2", Type: "int64"}}, Result: res}, true
    }
    if strings.HasSuffix(name, ".MetalDevices") || name == "gpu.MetalDevices" {
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "slice<Struct{ID:int64,Name:string,Backend:string}>"}
        return ir.Expr{Op: "call", Callee: "ami_rt_metal_devices", Result: res}, true
    }
    if strings.HasSuffix(name, ".CudaDevices") || name == "gpu.CudaDevices" {
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "slice<Struct{ID:int64,Name:string,Backend:string}>"}
        return ir.Expr{Op: "call", Callee: "ami_rt_cuda_devices", Result: res}, true
    }
    if strings.HasSuffix(name, ".OpenCLPlatforms") || name == "gpu.OpenCLPlatforms" {
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "slice<Struct{Name:string,Vendor:string}>"}
        return ir.Expr{Op: "call", Callee: "ami_rt_opencl_platforms", Result: res}, true
    }
    if strings.HasSuffix(name, ".OpenCLDevices") || name == "gpu.OpenCLDevices" {
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "slice<Struct{ID:int64,Name:string,Backend:string}>"}
        return ir.Expr{Op: "call", Callee: "ami_rt_opencl_devices", Result: res}, true
    }
    if strings.HasSuffix(name, ".MetalCreateContext") || name == "gpu.MetalCreateContext" {
        var args []ir.Value
        for _, a := range c.Args { if ex, ok := lowerExpr(st, a); ok && ex.Result != nil { args = append(args, *ex.Result) } }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "Owned"}
        return ir.Expr{Op: "call", Callee: "ami_rt_metal_ctx_create", Args: args, Result: res}, true
    }
    if strings.HasSuffix(name, ".MetalDestroyContext") || name == "gpu.MetalDestroyContext" {
        var args []ir.Value
        for _, a := range c.Args { if ex, ok := lowerExpr(st, a); ok && ex.Result != nil { args = append(args, *ex.Result) } }
        return ir.Expr{Op: "call", Callee: "ami_rt_metal_ctx_destroy", Args: args}, true
    }
    if strings.HasSuffix(name, ".MetalCompileLibrary") || name == "gpu.MetalCompileLibrary" {
        var args []ir.Value
        for _, a := range c.Args { if ex, ok := lowerExpr(st, a); ok && ex.Result != nil { args = append(args, *ex.Result) } }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "Owned"}
        return ir.Expr{Op: "call", Callee: "ami_rt_metal_lib_compile", Args: args, Result: res}, true
    }
    if strings.HasSuffix(name, ".MetalCreatePipeline") || name == "gpu.MetalCreatePipeline" {
        var args []ir.Value
        for _, a := range c.Args { if ex, ok := lowerExpr(st, a); ok && ex.Result != nil { args = append(args, *ex.Result) } }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "Owned"}
        return ir.Expr{Op: "call", Callee: "ami_rt_metal_pipe_create", Args: args, Result: res}, true
    }
    if strings.HasSuffix(name, ".MetalAlloc") || name == "gpu.MetalAlloc" {
        var args []ir.Value
        for _, a := range c.Args { if ex, ok := lowerExpr(st, a); ok && ex.Result != nil { args = append(args, *ex.Result) } }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "Owned"}
        return ir.Expr{Op: "call", Callee: "ami_rt_metal_alloc", Args: args, Result: res}, true
    }
    if strings.HasSuffix(name, ".MetalFree") || name == "gpu.MetalFree" {
        var args []ir.Value
        for _, a := range c.Args { if ex, ok := lowerExpr(st, a); ok && ex.Result != nil { args = append(args, *ex.Result) } }
        return ir.Expr{Op: "call", Callee: "ami_rt_metal_free", Args: args}, true
    }
    if strings.HasSuffix(name, ".MetalCopyToDevice") || name == "gpu.MetalCopyToDevice" {
        var args []ir.Value
        for _, a := range c.Args { if ex, ok := lowerExpr(st, a); ok && ex.Result != nil { args = append(args, *ex.Result) } }
        return ir.Expr{Op: "call", Callee: "ami_rt_metal_copy_to_device", Args: args}, true
    }
    if strings.HasSuffix(name, ".MetalCopyFromDevice") || name == "gpu.MetalCopyFromDevice" {
        var args []ir.Value
        for _, a := range c.Args { if ex, ok := lowerExpr(st, a); ok && ex.Result != nil { args = append(args, *ex.Result) } }
        return ir.Expr{Op: "call", Callee: "ami_rt_metal_copy_from_device", Args: args}, true
    }
    if strings.HasSuffix(name, ".MetalDispatchBlocking") || name == "gpu.MetalDispatchBlocking" {
        var args []ir.Value
        for _, a := range c.Args { if ex, ok := lowerExpr(st, a); ok && ex.Result != nil { args = append(args, *ex.Result) } }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "Error<any>"}
        return ir.Expr{Op: "call", Callee: "ami_rt_metal_dispatch_blocking", Args: args, Result: res}, true
    }
    if strings.HasSuffix(name, ".Sleep") || name == "time.Sleep" {
        var args []ir.Value
        for _, a := range c.Args { if ex, ok := lowerExpr(st, a); ok && ex.Result != nil { args = append(args, *ex.Result) } }
        return ir.Expr{Op: "call", Callee: "ami_rt_sleep_ms", Args: args}, true
    }
    if strings.HasSuffix(name, ".Now") || name == "time.Now" {
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "Time"}
        return ir.Expr{Op: "call", Callee: "ami_rt_time_now", Result: res}, true
    }
    if strings.HasSuffix(name, ".Add") || name == "time.Add" {
        var args []ir.Value
        if len(c.Args) >= 1 { if ex, ok := lowerExpr(st, c.Args[0]); ok && ex.Result != nil { args = append(args, ir.Value{ID: ex.Result.ID, Type: "int64"}) } }
        if len(c.Args) >= 2 { if ex, ok := lowerExpr(st, c.Args[1]); ok && ex.Result != nil { args = append(args, *ex.Result) } }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "Time"}
        return ir.Expr{Op: "call", Callee: "ami_rt_time_add", Args: args, Result: res}, true
    }
    if strings.HasSuffix(name, ".Delta") || name == "time.Delta" {
        var args []ir.Value
        for i := 0; i < len(c.Args) && i < 2; i++ { if ex, ok := lowerExpr(st, c.Args[i]); ok && ex.Result != nil { args = append(args, ir.Value{ID: ex.Result.ID, Type: "int64"}) } }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "int64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_time_delta", Args: args, Result: res}, true
    }
    if strings.HasSuffix(name, ".UnixNano") || name == "time.UnixNano" {
        var args []ir.Value
        if len(c.Args) >= 1 {
            if ex, ok := lowerExpr(st, c.Args[0]); ok && ex.Result != nil { args = append(args, ir.Value{ID: ex.Result.ID, Type: "int64"}) }
        } else {
            if recv, ok := synthesizeMethodRecvArg(st, c.Name); ok { args = append(args, recv) }
        }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "int64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_time_unix_nano", Args: args, Result: res}, true
    }
    if strings.HasSuffix(name, ".Unix") || name == "time.Unix" {
        var args []ir.Value
        if len(c.Args) >= 1 {
            if ex, ok := lowerExpr(st, c.Args[0]); ok && ex.Result != nil { args = append(args, ir.Value{ID: ex.Result.ID, Type: "int64"}) }
        } else {
            if recv, ok := synthesizeMethodRecvArg(st, c.Name); ok { args = append(args, recv) }
        }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "int64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_time_unix", Args: args, Result: res}, true
    }
    if (name == "signal.Enable") || (st != nil && st.funcParams != nil && len(st.funcParams[name]) == 1 && st.funcParams[name][0] == "SignalType") {
        var args []ir.Value
        if len(c.Args) >= 1 {
            switch s := c.Args[0].(type) {
            case *ast.SelectorExpr:
                var v int64
                switch s.Sel { case "SIGINT": v = 2; case "SIGTERM": v = 15; case "SIGHUP": v = 1; case "SIGQUIT": v = 3 }
                if v != 0 { args = append(args, ir.Value{ID: "#"+strconv.FormatInt(v, 10), Type: "int64"}) }
            }
            if len(args) == 0 { if ex, ok := lowerExpr(st, c.Args[0]); ok && ex.Result != nil { args = append(args, ir.Value{ID: ex.Result.ID, Type: "int64"}) } }
        }
        return ir.Expr{Op: "call", Callee: "ami_rt_os_signal_enable", Args: args}, true
    }
    if (name == "signal.Disable") || (st != nil && st.funcParams != nil && len(st.funcParams[name]) == 1 && st.funcParams[name][0] == "SignalType") {
        var args []ir.Value
        if len(c.Args) >= 1 {
            switch s := c.Args[0].(type) {
            case *ast.SelectorExpr:
                var v int64
                switch s.Sel { case "SIGINT": v = 2; case "SIGTERM": v = 15; case "SIGHUP": v = 1; case "SIGQUIT": v = 3 }
                if v != 0 { args = append(args, ir.Value{ID: "#"+strconv.FormatInt(v, 10), Type: "int64"}) }
            }
            if len(args) == 0 { if ex, ok := lowerExpr(st, c.Args[0]); ok && ex.Result != nil { args = append(args, ir.Value{ID: ex.Result.ID, Type: "int64"}) } }
        }
        return ir.Expr{Op: "call", Callee: "ami_rt_os_signal_disable", Args: args}, true
    }
    if (name == "signal.Install") || (st != nil && st.funcParams != nil && len(st.funcParams[name]) == 1 && st.funcParams[name][0] == "any" && (st.funcResults == nil || len(st.funcResults[name]) == 0)) {
        if len(c.Args) >= 1 {
            tok, ok := handlerTokenValue(c.Args[0])
            var args []ir.Value
            if ok { args = append(args, ir.Value{ID: "#"+strconv.FormatInt(tok, 10), Type: "int64"}) } else { args = append(args, ir.Value{ID: "#0", Type: "int64"}) }
            switch v := c.Args[0].(type) {
            case *ast.IdentExpr:
                if v.Name != "" { args = append(args, ir.Value{ID: "#@"+v.Name, Type: "ptr"}) } else { args = append(args, ir.Value{ID: "#null", Type: "ptr"}) }
            default:
                args = append(args, ir.Value{ID: "#null", Type: "ptr"})
            }
            return ir.Expr{Op: "call", Callee: "ami_rt_install_handler_thunk", Args: args}, true
        }
        return ir.Expr{Op: "call", Callee: "ami_rt_install_handler_thunk", Args: []ir.Value{{ID: "#0", Type: "int64"}, {ID: "#null", Type: "ptr"}}}, true
    }
    if (name == "signal.Token") || (st != nil && st.funcParams != nil && len(st.funcParams[name]) == 1 && st.funcParams[name][0] == "any" && st.funcResults != nil && len(st.funcResults[name]) >= 1 && st.funcResults[name][0] == "int64") {
        if len(c.Args) >= 1 {
            tok, ok := handlerTokenValue(c.Args[0])
            id := st.newTemp(); res := &ir.Value{ID: id, Type: "int64"}
            if ok { return ir.Expr{Op: "lit:" + strconv.FormatInt(tok, 10), Result: res}, true }
            return ir.Expr{Op: "lit:0", Result: res}, true
        }
        id := st.newTemp(); res := &ir.Value{ID: id, Type: "int64"}
        return ir.Expr{Op: "lit:0", Result: res}, true
    }
    return ir.Expr{}, false
}
