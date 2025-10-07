package llvm

import (
    "strings"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// emitWorkerCores scans IR functions to find worker-shaped functions and emits
// JSON-ABI core wrappers:
//   i8* @ami_worker_core_<name>(i8* in, i32 inlen, i32* outlen, i8** err)
// For now, wrappers return an "unimplemented" error string to establish the ABI.
func emitWorkerCores(e *ModuleEmitter, fns []ir.Function) {
    // Ensure required externs are present
    e.RequireExtern("declare ptr @malloc(i64)")
    e.RequireExtern("declare void @llvm.memcpy.p0.p0.i64(ptr, ptr, i64, i1)")
    for _, fn := range fns {
        if len(fn.Params) != 1 || len(fn.Results) != 2 { continue }
        p0 := fn.Params[0].Type
        r1 := fn.Results[1].Type
        if !strings.HasPrefix(p0, "Event<") || r1 != "error" { continue }
        name := sanitizeIdent(fn.Name)
        // emit private constant for error message (with NUL)
        msg := "unimplemented"
        esc := encodeCString(msg)
        n := len(msg) + 1
        gname := "@.wcore.err." + name
        e.AddGlobal(gname + " = private constant [" + itoa(n) + " x i8] c\"" + esc + "\"")
        // function body: allocate and copy message, set *err, return null
        var b strings.Builder
        b.WriteString("define i8* @ami_worker_core_")
        b.WriteString(name)
        b.WriteString("(i8* %in, i32 %inlen, i32* %outlen, i8** %err) {\n")
        b.WriteString("entry:\n")
        b.WriteString("  %msg = getelementptr inbounds [")
        b.WriteString(itoa(n))
        b.WriteString(" x i8], ptr ")
        b.WriteString(gname)
        b.WriteString(", i64 0, i64 0\n")
        b.WriteString("  %sz = zext i32 ")
        b.WriteString(itoa(n))
        b.WriteString(" to i64\n")
        b.WriteString("  %buf = call ptr @malloc(i64 %sz)\n")
        b.WriteString("  call void @llvm.memcpy.p0.p0.i64(ptr %buf, ptr %msg, i64 %sz, i1 false)\n")
        b.WriteString("  store ptr %buf, ptr %err, align 8\n")
        b.WriteString("  ret ptr null\n}\n")
        e.funcs = append(e.funcs, b.String())
    }
}

