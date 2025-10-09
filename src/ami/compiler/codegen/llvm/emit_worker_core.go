package llvm

import (
    "sort"
    "strings"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    types "github.com/sam-caldwell/ami/src/ami/compiler/types"
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
    e.RequireExtern("declare ptr @ami_rt_structured_to_json(ptr, i32*)")
    e.RequireExtern("declare ptr @ami_rt_value_to_json(ptr, i64, i32*)")
    e.RequireExtern("declare ptr @ami_rt_error_to_cstring(ptr, i32*)")
    // Primitive payload JSON bridges
    e.RequireExtern("declare ptr @ami_rt_bool_to_json(i1, i32*)")
    e.RequireExtern("declare ptr @ami_rt_i64_to_json(i64, i32*)")
    e.RequireExtern("declare ptr @ami_rt_double_to_json(double, i32*)")
    e.RequireExtern("declare ptr @ami_rt_string_to_json(ptr, i32*)")
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
        // Use ABI-safe types for call/result aggregate
        ty0 := abiType(fn.Results[0].Type)
        ty1 := abiType(fn.Results[1].Type)
        // function body
        var b strings.Builder
        b.WriteString("define i8* @ami_worker_core_")
        b.WriteString(name)
        b.WriteString("(i8* %in, i32 %inlen, i32* %outlen, i8** %err) {\n")
        b.WriteString("entry:\n")
        // Convert JSON input to Event handle
        b.WriteString("  %ev = call ptr @ami_rt_json_to_event(ptr %in, i32 %inlen)\n")
        // Convert ev handle to ABI param (i64) and call worker function: {ty0, ty1} @name(<abi ev>)
        b.WriteString("  %ev_i64 = ptrtoint ptr %ev to i64\n")
        b.WriteString("  %res = call { ")
        b.WriteString(ty0); b.WriteString(", "); b.WriteString(ty1)
        b.WriteString(" } @")
        b.WriteString(fn.Name)
        // Param type for Event<...> at ABI is i64
        b.WriteString("(i64 %ev_i64)\n")
        b.WriteString("  %er = extractvalue { "); b.WriteString(ty0); b.WriteString(", "); b.WriteString(ty1); b.WriteString(" } %res, 1\n")
        // Check error handle is zero (ABI i64); treat 0 as null
        if ty1 == "i64" {
            b.WriteString("  %ok = icmp eq i64 %er, 0\n")
        } else {
            b.WriteString("  %ok = icmp eq "); b.WriteString(ty1); b.WriteString(" %er, null\n")
        }
        b.WriteString("  br i1 %ok, label %oklbl, label %errlbl\n")
        // Error path: convert error handle to cstring and return null with *err set
        b.WriteString("errlbl:\n")
        if ty1 == "i64" {
            b.WriteString("  %er_ptr = inttoptr i64 %er to ptr\n")
            b.WriteString("  %cerr = call ptr @ami_rt_error_to_cstring(ptr %er_ptr, i32* %outlen)\n")
        } else {
            b.WriteString("  %cerr = call ptr @ami_rt_error_to_cstring(ptr %er, i32* %outlen)\n")
        }
        b.WriteString("  store ptr %cerr, ptr %err, align 8\n")
        b.WriteString("  ret ptr null\n")
        // Success path: emit JSON based on first result type
        b.WriteString("oklbl:\n")
        if strings.HasPrefix(fn.Results[0].Type, "Event<") {
            b.WriteString("  %r0 = extractvalue { "); b.WriteString(ty0); b.WriteString(", "); b.WriteString(ty1); b.WriteString(" } %res, 0\n")
            // Convert ABI handle back to ptr for JSON bridge when needed
            if ty0 == "i64" {
                b.WriteString("  %r0_ptr = inttoptr i64 %r0 to ptr\n")
                b.WriteString("  %js = call ptr @ami_rt_event_to_json(ptr %r0_ptr, i32* %outlen)\n")
            } else {
                b.WriteString("  %js = call ptr @ami_rt_event_to_json(ptr %r0, i32* %outlen)\n")
            }
            b.WriteString("  store ptr null, ptr %err, align 8\n")
            b.WriteString("  ret ptr %js\n")
        } else {
            // Bare payload: convert to JSON via primitive/string bridges when possible; fallback to null
            b.WriteString("  %r0 = extractvalue { "); b.WriteString(ty0); b.WriteString(", "); b.WriteString(ty1); b.WriteString(" } %res, 0\n")
            // Handle string textual type specially (ABI i64 handle -> ptr)
            if strings.TrimSpace(fn.Results[0].Type) == "string" {
                if ty0 == "i64" {
                    b.WriteString("  %r0_ptr = inttoptr i64 %r0 to ptr\n")
                    b.WriteString("  %js = call ptr @ami_rt_string_to_json(ptr %r0_ptr, i32* %outlen)\n")
                } else {
                    b.WriteString("  %js = call ptr @ami_rt_string_to_json(ptr %r0, i32* %outlen)\n")
                }
            } else if strings.Contains(fn.Results[0].Type, "Struct{") || strings.HasPrefix(strings.TrimSpace(fn.Results[0].Type), "slice<") {
                // Structured: build a compact binary descriptor: 'S'|'A' + i32 count + entries
                // For Struct: entries are (u8 nameLen, nameBytes...) in sorted field order.
                tdBytes := make([]byte, 0, 64)
                tstr := strings.TrimSpace(fn.Results[0].Type)
                if strings.HasPrefix(tstr, "Struct{") {
                    tdBytes = append(tdBytes, 'S')
                    if tt, err := types.Parse(tstr); err == nil {
                        if s, ok := tt.(types.Struct); ok {
                            // Collect field names sorted
                            names := make([]string, 0, len(s.Fields))
                            for k := range s.Fields { names = append(names, k) }
                            sort.Strings(names)
                            // count as little-endian i32
                            cnt := len(names)
                            tdBytes = append(tdBytes, byte(cnt), byte(cnt>>8), byte(cnt>>16), byte(cnt>>24))
                            for _, name := range names {
                                if len(name) > 255 { name = name[:255] }
                                tdBytes = append(tdBytes, byte(len(name)))
                                tdBytes = append(tdBytes, []byte(name)...)
                                // kind byte per field (i=int64, d=double, b=bool, s=string, o=Owned/default)
                                kind := byte('o')
                                if ft, ok2 := s.Fields[name]; ok2 {
                                    switch ft := ft.(type) {
                                    case types.Primitive:
                                        switch ft.K {
                                        case types.Int, types.Int64:
                                            kind = 'i'
                                        case types.Float64:
                                            kind = 'd'
                                        case types.Bool:
                                            kind = 'b'
                                        case types.String:
                                            kind = 's'
                                        }
                                    default:
                                        // leave as 'o'
                                    }
                                }
                                tdBytes = append(tdBytes, kind)
                            }
                        } else {
                            // unknown parse: encode zero-count struct
                            tdBytes = append(tdBytes, 0, 0, 0, 0)
                        }
                    } else {
                        tdBytes = append(tdBytes, 0, 0, 0, 0)
                    }
                } else {
                    // slice<...>: unknown length at compile-time → encode 'A' + count=0 + element kind [+ nested struct descriptor]
                    tdBytes = append(tdBytes, 'A', 0, 0, 0, 0)
                    kind := byte('o')
                    if tt, err := types.Parse(tstr); err == nil {
                        // Expect Generic slice
                        switch st := tt.(type) {
                        case types.Generic:
                            if strings.EqualFold(st.Name, "slice") && len(st.Args) == 1 {
                                switch et := st.Args[0].(type) {
                                case types.Primitive:
                                    switch et.K {
                                    case types.Int, types.Int64:
                                        kind = 'i'
                                    case types.Float64:
                                        kind = 'd'
                                    case types.Bool:
                                        kind = 'b'
                                    case types.String:
                                        kind = 's'
                                    }
                                case types.Struct:
                                    kind = 'S'
                                    // inline nested struct descriptor after elem kind
                                    // Build nested struct descriptor: 'S' + count + entries(nameLen,nameBytes,kind)
                                    b := []byte{'S'}
                                    // collect sorted names
                                    names := make([]string, 0, len(et.Fields))
                                    for k := range et.Fields { names = append(names, k) }
                                    sort.Strings(names)
                                    cnt := len(names)
                                    b = append(b, byte(cnt), byte(cnt>>8), byte(cnt>>16), byte(cnt>>24))
                                    for _, name := range names {
                                        if len(name) > 255 { name = name[:255] }
                                        b = append(b, byte(len(name)))
                                        b = append(b, []byte(name)...)
                                        // field kind for nested struct
                                        fk := byte('o')
                                        if ft, ok2 := et.Fields[name]; ok2 {
                                            switch ft := ft.(type) {
                                            case types.Primitive:
                                                switch ft.K {
                                                case types.Int, types.Int64:
                                                    fk = 'i'
                                                case types.Float64:
                                                    fk = 'd'
                                                case types.Bool:
                                                    fk = 'b'
                                                case types.String:
                                                    fk = 's'
                                                }
                                            }
                                        }
                                        b = append(b, fk)
                                    }
                                    tdBytes = append(tdBytes, kind)
                                    tdBytes = append(tdBytes, b...)
                                    kind = 0 // already appended kind and nested descriptor
                                }
                            }
                        }
                    }
                    if kind != 0 { tdBytes = append(tdBytes, kind) }
                }
                tdStr := string(tdBytes)
                esc := encodeCString(tdStr)
                n := len(tdStr) + 1
                gname := "@.td." + sanitizeIdent(fn.Name) + "." + sanitizeIdent(fn.Results[0].Type)
                e.AddGlobal(gname + " = private constant [" + itoa(n) + " x i8] c\"" + esc + "\"")
                b.WriteString("  %td = getelementptr inbounds ["); b.WriteString(itoa(n)); b.WriteString(" x i8], ptr ")
                b.WriteString(gname); b.WriteString(", i64 0, i64 0\n")
                if ty0 == "i64" {
                    b.WriteString("  %js = call ptr @ami_rt_value_to_json(ptr %td, i64 %r0, i32* %outlen)\n")
                } else {
                    b.WriteString("  %r0_i64 = ptrtoint ptr %r0 to i64\n")
                    b.WriteString("  %js = call ptr @ami_rt_value_to_json(ptr %td, i64 %r0_i64, i32* %outlen)\n")
                }
            } else if ty0 == "i1" {
                b.WriteString("  %js = call ptr @ami_rt_bool_to_json(i1 %r0, i32* %outlen)\n")
            } else if ty0 == "i64" {
                b.WriteString("  %js = call ptr @ami_rt_i64_to_json(i64 %r0, i32* %outlen)\n")
            } else if ty0 == "double" {
                b.WriteString("  %js = call ptr @ami_rt_double_to_json(double %r0, i32* %outlen)\n")
            } else {
                b.WriteString("  %js = call ptr @ami_rt_payload_to_json(ptr null, i32* %outlen)\n")
            }
            b.WriteString("  store ptr null, ptr %err, align 8\n")
            b.WriteString("  ret ptr %js\n")
        }
        b.WriteString("}\n")
        e.funcs = append(e.funcs, b.String())
    }
}
