package llvm

import (
    "strings"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// emitWorkerCores scans IR functions to find worker-shaped functions and emits
// JSON-ABI core wrappers:
//   i8* @ami_worker_core_<name>(i8* in, i32 inlen, i32* outlen, i8** err)
// For now, wrappers convert input JSON to an Event handle, call the worker, and
// on success convert the first result to JSON. When the first result is not an
// Event<...>, wrappers return an "unimplemented" error string (payload JSON to follow).
func emitWorkerCores(e *ModuleEmitter, fns []ir.Function) {
    // Ensure required externs are present
    e.RequireExtern("declare ptr @malloc(i64)")
    e.RequireExtern("declare void @llvm.memcpy.p0.p0.i64(ptr, ptr, i64, i1)")
    e.RequireExtern("declare ptr @ami_rt_json_to_event(ptr, i32)")
    e.RequireExtern("declare ptr @ami_rt_event_to_json(ptr, i32*)")
    e.RequireExtern("declare ptr @ami_rt_payload_to_json(ptr, i32*)")
    for _, fn := range fns {
        if len(fn.Params) != 1 || len(fn.Results) != 2 { continue }
        p0 := fn.Params[0].Type
        r1 := fn.Results[1].Type
        if !strings.HasPrefix(p0, "Event<") || r1 != "error" { continue }
        name := sanitizeIdent(fn.Name)
        // emit private constant messages (with NUL)
        msg := "unimplemented"
        esc := encodeCString(msg)
        n := len(msg) + 1
        gname := "@.wcore.err." + name
        e.AddGlobal(gname + " = private constant [" + itoa(n) + " x i8] c\"" + esc + "\"")
        msg2 := "worker error"
        esc2 := encodeCString(msg2)
        n2 := len(msg2) + 1
        gname2 := "@.wcore.werr." + name
        e.AddGlobal(gname2 + " = private constant [" + itoa(n2) + " x i8] c\"" + esc2 + "\"")
        // Prepare types
        ty0 := mapType(fn.Results[0].Type)
        ty1 := mapType(fn.Results[1].Type) // error -> ptr
        // function body
        var b strings.Builder
        b.WriteString("define i8* @ami_worker_core_")
        b.WriteString(name)
        b.WriteString("(i8* %in, i32 %inlen, i32* %outlen, i8** %err) {\n")
        b.WriteString("entry:\n")
        // Convert JSON input to Event handle
        b.WriteString("  %ev = call ptr @ami_rt_json_to_event(ptr %in, i32 %inlen)\n")
        // Call worker function: {ty0, ty1} @name(ptr %ev)
        b.WriteString("  %res = call { ")
        b.WriteString(ty0); b.WriteString(", "); b.WriteString(ty1)
        b.WriteString(" } @")
        b.WriteString(fn.Name)
        b.WriteString("(ptr %ev)\n")
        b.WriteString("  %er = extractvalue { "); b.WriteString(ty0); b.WriteString(", "); b.WriteString(ty1); b.WriteString(" } %res, 1\n")
        b.WriteString("  %ok = icmp eq ptr %er, null\n")
        b.WriteString("  br i1 %ok, label %oklbl, label %errlbl\n")
        // Error path: set *err to "worker error" and return null
        b.WriteString("errlbl:\n")
        b.WriteString("  %msgp2 = getelementptr inbounds [")
        b.WriteString(itoa(n2))
        b.WriteString(" x i8], ptr ")
        b.WriteString(gname2)
        b.WriteString(", i64 0, i64 0\n")
        b.WriteString("  %sz2 = zext i32 ")
        b.WriteString(itoa(n2))
        b.WriteString(" to i64\n")
        b.WriteString("  %buf2 = call ptr @malloc(i64 %sz2)\n")
        b.WriteString("  call void @llvm.memcpy.p0.p0.i64(ptr %buf2, ptr %msgp2, i64 %sz2, i1 false)\n")
        b.WriteString("  store ptr %buf2, ptr %err, align 8\n")
        b.WriteString("  ret ptr null\n")
        // Success path: emit JSON based on first result type
        b.WriteString("oklbl:\n")
        if strings.HasPrefix(fn.Results[0].Type, "Event<") {
            b.WriteString("  %r0 = extractvalue { "); b.WriteString(ty0); b.WriteString(", "); b.WriteString(ty1); b.WriteString(" } %res, 0\n")
            b.WriteString("  %js = call ptr @ami_rt_event_to_json(ptr %r0, i32* %outlen)\n")
            b.WriteString("  store ptr null, ptr %err, align 8\n")
            b.WriteString("  ret ptr %js\n")
        } else {
            // Bare payload: convert to JSON via payload bridge (bring-up returns "null")
            b.WriteString("  %js = call ptr @ami_rt_payload_to_json(ptr null, i32* %outlen)\n")
            b.WriteString("  store ptr null, ptr %err, align 8\n")
            b.WriteString("  ret ptr %js\n")
        }
        b.WriteString("}\n")
        e.funcs = append(e.funcs, b.String())
    }
}
