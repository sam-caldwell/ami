package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "time"
    "io"

    "github.com/sam-caldwell/ami/src/ami/compiler/codegen"
    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// buildLink performs the link stage per env, caching runtime.o under build/runtime/<env>/.
// Emits diagnostics when jsonOut is true and continues best-effort.
func buildLink(out io.Writer, dir string, ws workspace.Workspace, envs []string, jsonOut bool) {
    be := codegen.DefaultBackend()
    clang, ferr := be.FindToolchain()
    if ferr != nil {
        if lg := getRootLogger(); lg != nil { lg.Info("build.toolchain.missing", map[string]any{"tool": "clang"}) }
        return
    }
    // Resolve binary name
    binName := "app"
    if mp := ws.FindPackage("main"); mp != nil && mp.Name != "" { binName = mp.Name }
    // Per-env link
    for _, env := range envs {
        if containsEnv(buildNoLinkEnvs, env) { continue }
        // collect per-env objects
        var objs []string
        for _, e := range ws.Packages {
            glob := filepath.Join(dir, "build", env, "obj", e.Package.Name, "*.o")
            if matches, _ := filepath.Glob(glob); len(matches) > 0 { objs = append(objs, matches...) }
        }
        if len(objs) == 0 { continue }
        triple := be.TripleForEnv(env)
        rtDir := filepath.Join(dir, "build", "runtime", env)
        rtObj := filepath.Join(rtDir, "runtime.o")
        if _, stErr := os.Stat(rtObj); stErr != nil {
            if llPath, werr := be.WriteRuntimeLL(rtDir, triple, false); werr == nil {
                if cerr := be.CompileLLToObject(clang, llPath, rtObj, triple); cerr != nil {
                    if lg := getRootLogger(); lg != nil { lg.Info("build.runtime.obj.fail", map[string]any{"error": cerr.Error(), "env": env}) }
                    if jsonOut {
                        d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_OBJ_COMPILE_FAIL", Message: "failed to compile LLVM to object (runtime)", File: llPath, Data: map[string]any{"env": env, "what": "runtime"}}
                        if te, ok := cerr.(llvme.ToolError); ok { if d.Data == nil { d.Data = map[string]any{} }; d.Data["stderr"] = te.Stderr }
                        _ = json.NewEncoder(out).Encode(d)
                    }
                }
            } else {
                if lg := getRootLogger(); lg != nil { lg.Info("build.runtime.ll.fail", map[string]any{"error": werr.Error(), "env": env}) }
                if jsonOut {
                    d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_LLVM_EMIT", Message: werr.Error(), File: filepath.Join(rtDir, "runtime.ll"), Data: map[string]any{"env": env}}
                    _ = json.NewEncoder(out).Encode(d)
                }
            }
        }
        if st, _ := os.Stat(rtObj); st != nil { objs = append(objs, rtObj) }
        // optional entry.o when no user main
        if !hasUserMain(ws, dir) {
            ingress := collectIngressIDs(ws, dir)
            if entLL, eerr := be.WriteIngressEntrypointLL(rtDir, triple, ingress); eerr == nil {
                entObj := filepath.Join(rtDir, "entry.o")
                if ecomp := be.CompileLLToObject(clang, entLL, entObj, triple); ecomp == nil {
                    objs = append(objs, entObj)
                } else if jsonOut {
                    d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_OBJ_COMPILE_FAIL", Message: "failed to compile LLVM to object (entry)", File: entLL, Data: map[string]any{"env": env, "what": "entry"}}
                    if te, ok := ecomp.(llvme.ToolError); ok { if d.Data == nil { d.Data = map[string]any{} }; d.Data["stderr"] = te.Stderr }
                    _ = json.NewEncoder(out).Encode(d)
                }
            }
        }
        outDir := filepath.Join(dir, "build", env)
        _ = os.MkdirAll(outDir, 0o755)
        outBin := filepath.Join(outDir, binName)
        extra := linkExtraFlags(env, ws.Toolchain.Linker.Options)
        if lerr := be.LinkObjects(clang, objs, outBin, triple, extra...); lerr != nil {
            if lg := getRootLogger(); lg != nil { lg.Info("build.link.fail", map[string]any{"error": lerr.Error(), "bin": outBin, "env": env}) }
            if jsonOut {
                data := map[string]any{"env": env, "bin": outBin}
                if te, ok := lerr.(llvme.ToolError); ok { data["stderr"] = te.Stderr }
                d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_LINK_FAIL", Message: "linking failed", File: "clang", Data: data}
                _ = json.NewEncoder(out).Encode(d)
            }
        } else if lg := getRootLogger(); lg != nil {
            lg.Info("build.link.ok", map[string]any{"bin": outBin, "objects": len(objs), "env": env})
        }
    }
}
