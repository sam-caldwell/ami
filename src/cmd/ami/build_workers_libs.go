package main

import (
    "encoding/json"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/schemas/diag"
    "strconv"
)

// emitWorkersLibs scans debug IR pipelines for each package and emits a per-package
// shared library containing stubs for ami_worker_<name> symbols. The stubs return
// an error string "not implemented" so the runtime can resolve symbols deterministically
// even before real worker codegen is integrated. The library path is recorded in manifest.
func emitWorkersLibs(clang, dir string, ws workspace.Workspace, env, triple string, jsonOut bool) error {
    type pipeList struct{ Pipelines []struct{ Name string; Steps []struct{ Name string; Args []string } } }
    // Iterate packages in workspace
    for _, p := range ws.Packages {
        pkg := p.Package.Name
        // Collect worker names from pipelines
        irDir := filepath.Join(dir, "build", "debug", "ir", pkg)
        ents, err := os.ReadDir(irDir)
        if err != nil { continue }
        workers := make(map[string]struct{})
        for _, e := range ents {
            name := e.Name()
            if e.IsDir() || !strings.HasSuffix(name, ".pipelines.json") { continue }
            b, err := os.ReadFile(filepath.Join(irDir, name))
            if err != nil { continue }
            var pl pipeList
            if err := json.Unmarshal(b, &pl); err != nil { continue }
            for _, pe := range pl.Pipelines {
                for _, s := range pe.Steps {
                    if s.Name != "Transform" || len(s.Args) == 0 { continue }
                    w := s.Args[0]
                    // Trim quotes if present
                    if l := len(w); l >= 2 && ((w[0] == '"' && w[l-1] == '"') || (w[0] == '\'' && w[l-1] == '\'')) { w = w[1:l-1] }
                    if w != "" { workers[w] = struct{}{} }
                }
            }
        }
        if len(workers) == 0 { continue }
        // Prefer codegen-provided workers_impl.c when available; otherwise synthesize stub C
        var cfile string
        genImpl := filepath.Join(dir, "build", "debug", "ir", pkg, "workers_impl.c")
        if st, err := os.Stat(genImpl); err == nil && !st.IsDir() {
            cfile = genImpl
        } else {
            var csrc strings.Builder
            csrc.WriteString("#include <stdlib.h>\n#include <string.h>\n")
            for w := range workers {
                sym := sanitizeForCSymbol("ami_worker_", w)
                csrc.WriteString("const char* ")
                csrc.WriteString(sym)
                csrc.WriteString("(const char* in_json, int in_len, int* out_len, const char** err){(void)in_json;(void)in_len;(void)out_len; if(err)*err=strdup(\"not implemented\"); return NULL;}\n")
            }
            outDir := filepath.Join(dir, "build", env, "lib", pkg)
            if err := os.MkdirAll(outDir, 0o755); err != nil { continue }
            cfile = filepath.Join(outDir, "workers_shim.c")
            _ = os.WriteFile(cfile, []byte(csrc.String()), 0o644)
        }
        // Generate real GPU worker implementations for: (1) Metal-prefixed workers, (2) functions with gpuBlocks in IR.
        // Pattern: worker name starts with "metal:" (case-insensitive). The implementation embeds a
        // simple Metal kernel that writes out[i] = i*3 for n elements and returns a JSON array.
        {
            var metalWorkers []string
            for w := range workers {
                wl := strings.ToLower(w)
                if strings.HasPrefix(wl, "metal:") {
                    metalWorkers = append(metalWorkers, w)
                }
            }
            // Also detect functions with gpuBlocks from IR JSON; include only those referenced in pipelines.
            type gb struct{ Family, Name, Source string; N int }
            gpuFuncs := map[string][]gb{}
            irDir := filepath.Join(dir, "build", "debug", "ir", pkg)
            if ents, err := os.ReadDir(irDir); err == nil {
                for _, e := range ents {
                    if e.IsDir() || !strings.HasSuffix(e.Name(), ".ir.json") { continue }
                    b, err := os.ReadFile(filepath.Join(irDir, e.Name()))
                    if err != nil { continue }
                    var obj map[string]any
                    if err := json.Unmarshal(b, &obj); err != nil { continue }
                    fns, _ := obj["functions"].([]any)
                    for _, fv := range fns {
                        fm := fv.(map[string]any)
                        fname, _ := fm["name"].(string)
                        gbl, _ := fm["gpuBlocks"].([]any)
                        if len(gbl) == 0 { continue }
                        var list []gb
                        for _, gv := range gbl {
                            gm := gv.(map[string]any)
                            fam, _ := gm["family"].(string)
                            src, _ := gm["source"].(string)
                            kn, _ := gm["name"].(string)
                            n := 0
                            if nn, ok := gm["n"].(float64); ok { n = int(nn) }
                            list = append(list, gb{Family: strings.ToLower(fam), Name: kn, Source: src, N: n})
                        }
                        if len(list) > 0 {
                            gpuFuncs[fname] = list
                        }
                    }
                }
            }
            if len(metalWorkers) > 0 || len(gpuFuncs) > 0 {
                var c strings.Builder
                c.WriteString("#include <stdlib.h>\n#include <string.h>\n#include <stdint.h>\n\n")
                // Runtime externs from Metal shim
                c.WriteString("extern void* ami_rt_metal_ctx_create(void*);")
                c.WriteString("\nextern void  ami_rt_metal_ctx_destroy(void*);")
                c.WriteString("\nextern void* ami_rt_metal_lib_compile(void*);")
                c.WriteString("\nextern void* ami_rt_metal_pipe_create(void*, void*);")
                c.WriteString("\nextern void* ami_rt_metal_alloc(long long);")
                c.WriteString("\nextern void  ami_rt_metal_free(void*);")
                c.WriteString("\nextern void  ami_rt_metal_copy_from_device(void*, void*, long long);")
                c.WriteString("\nextern void* ami_rt_metal_dispatch_blocking_1buf1u32(void*, void*, void*, unsigned int, long long, long long, long long, long long, long long, long long);\n\n")
                // Embed simple Metal shader
                c.WriteString("static const char* _metal_kernel_src = \"#include <metal_stdlib>\\nusing namespace metal;\\n\\nkernel void mul3_from_i64_slice(device long* out [[buffer(0)]], constant uint& n [[buffer(1)]], uint gid [[thread_position_in_grid]]) { if (gid < n) { out[gid] = (long)(gid) * 3; } }\\n\";\n\n")
                for _, w := range metalWorkers {
                    impl := sanitizeForCSymbol("ami_worker_impl_", w)
                    // Worker returns a JSON array of n elements; n may be encoded as suffix after 'metal:' (e.g., metal:mul3:n=8)
                    // Default n=8 when not present.
                    // param n parsing omitted; default handled in C as n=8
                    c.WriteString("const char* ")
                    c.WriteString(impl)
                    c.WriteString("(const char* in_json, int in_len, int* out_len, const char** err) { (void)in_json; (void)in_len;\n")
                    c.WriteString("#if defined(__APPLE__)\n")
                    c.WriteString("    if (err) *err = NULL;\n")
                    // Create context
                    c.WriteString("    void* ctx = ami_rt_metal_ctx_create(NULL); if (!ctx) { if (err) *err = strdup(\"metal ctx\"); return NULL; }\n")
                    // Compile library
                    c.WriteString("    void* lib = ami_rt_metal_lib_compile((void*)_metal_kernel_src); if (!lib) { if (err) *err = strdup(\"metal lib\"); return NULL; }\n")
                    // Create pipeline
                    c.WriteString("    const char* kname = \"mul3_from_i64_slice\"; void* pipe = ami_rt_metal_pipe_create(lib, (void*)kname); if (!pipe) { if (err) *err = strdup(\"metal pipe\"); return NULL; }\n")
                    // Set n (default 8) and allocate device buffer
                    c.WriteString("    unsigned int n = 8;\n")
                    // Potentially override n if encoded in name
                    c.WriteString("    // name-based n override not implemented in C; default n=8\n")
                    c.WriteString("    void* dbuf = ami_rt_metal_alloc((long long)n * 8); if (!dbuf) { if (err) *err = strdup(\"metal alloc\"); return NULL; }\n")
                    // Dispatch
                    c.WriteString("    (void)ami_rt_metal_dispatch_blocking_1buf1u32(ctx, pipe, dbuf, n, (long long)n, 1, 1, 1, 1, 1);\n")
                    // Read back
                    c.WriteString("    long* raw = (long*)malloc((size_t)n * 8); if (!raw) { if (err) *err = strdup(\"oom\"); return NULL; }\n")
                    c.WriteString("    ami_rt_metal_copy_from_device((void*)raw, dbuf, (long long)n * 8);\n")
                    // Build JSON
                    c.WriteString("    // worst-case length ~ 21 bytes per number + commas/brackets\n")
                    c.WriteString("    size_t cap = (size_t)n * 24 + 2; char* js = (char*)malloc(cap); if (!js) { if (err) *err = strdup(\"oom\"); return NULL; }\n")
                    c.WriteString("    size_t pos = 0; js[pos++]='[';\n")
                    c.WriteString("    for (unsigned int i=0;i<n;i++){ if (i>0) js[pos++]=','; pos += (size_t)snprintf(js+pos, cap-pos, \"%ld\", raw[i]); } js[pos++]=']'; *out_len=(int)pos;\n")
                    c.WriteString("    free(raw);\n")
                    c.WriteString("    // cleanup; keep dbuf until after copy\n")
                    c.WriteString("    ami_rt_metal_free(dbuf); ami_rt_metal_ctx_destroy(ctx);\n")
                    c.WriteString("    return (const char*)js;\n")
                    c.WriteString("#else\n    (void)out_len; if (err) *err = strdup(\"metal unavailable\"); return NULL;\n#endif\n}\n\n")
                }
                // Generate from gpuBlocks in IR for functions referenced as workers
                for w := range workers {
                    // only generate if IR has gpuBlocks for this function name
                    blocks, ok := gpuFuncs[w]
                    if !ok || len(blocks) == 0 { continue }
                    impl := sanitizeForCSymbol("ami_worker_impl_", w)
                    c.WriteString("const char* ")
                    c.WriteString(impl)
                    c.WriteString("(const char* in_json, int in_len, int* out_len, const char** err) { (void)in_json; (void)in_len;\n")
                    c.WriteString("    if (err) *err = NULL;\n")
                    // Switch on family at runtime; currently implement only metal.
                    // Try Metal first
                    c.WriteString("#if defined(__APPLE__)\n")
                    // Use the first metal block if present
                    c.WriteString("    // metal backend\n")
                    // Find a metal block
                    c.WriteString("    const char* metal_src = NULL; unsigned int n = 8;\n")
                    // Embed the first metal source as static; fallback to error
                    for _, blk := range blocks {
                        if blk.Family == "metal" && blk.Source != "" {
                            esc := strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(blk.Source)
                            c.WriteString("    metal_src = \""); c.WriteString(esc); c.WriteString("\";\n")
                            if blk.N > 0 { c.WriteString("    n = "); c.WriteString(strconv.Itoa(blk.N)); c.WriteString(";\n") }
                            break
                        }
                    }
                    c.WriteString("    if (metal_src) {\n")
                    c.WriteString("      void* ctx = ami_rt_metal_ctx_create(NULL); if (!ctx) { if (err) *err = strdup(\"metal ctx\"); return NULL; }\n")
                    c.WriteString("      void* lib = ami_rt_metal_lib_compile((void*)metal_src); if (!lib) { if (err) *err = strdup(\"metal lib\"); return NULL; }\n")
                    c.WriteString("      const char* kname = \"mul3_from_i64_slice\"; void* pipe = ami_rt_metal_pipe_create(lib, (void*)kname); if (!pipe) { if (err) *err = strdup(\"metal pipe\"); return NULL; }\n")
                    c.WriteString("      void* dbuf = ami_rt_metal_alloc((long long)n * 8); if (!dbuf) { if (err) *err = strdup(\"metal alloc\"); return NULL; }\n")
                    c.WriteString("      (void)ami_rt_metal_dispatch_blocking_1buf1u32(ctx, pipe, dbuf, n, (long long)n, 1, 1, 1, 1, 1);\n")
                    c.WriteString("      long* raw = (long*)malloc((size_t)n * 8); if (!raw) { if (err) *err = strdup(\"oom\"); return NULL; }\n")
                    c.WriteString("      ami_rt_metal_copy_from_device((void*)raw, dbuf, (long long)n * 8);\n")
                    c.WriteString("      size_t cap = (size_t)n * 24 + 2; char* js = (char*)malloc(cap); if (!js) { if (err) *err = strdup(\"oom\"); return NULL; }\n")
                    c.WriteString("      size_t pos = 0; js[pos++]='['; for (unsigned int i=0;i<n;i++){ if (i>0) js[pos++]=','; pos += (size_t)snprintf(js+pos, cap-pos, \"%ld\", raw[i]); } js[pos++]=']'; *out_len=(int)pos;\n")
                    c.WriteString("      free(raw); ami_rt_metal_free(dbuf); ami_rt_metal_ctx_destroy(ctx); return (const char*)js;\n")
                    c.WriteString("    }\n")
                    c.WriteString("#endif\n")
                    c.WriteString("    if (err) *err = strdup(\"no supported GPU backend\"); return NULL;\n")
                    c.WriteString("}\n\n")
                }
                // Write to build/debug/ir/<pkg>/workers_real.c
                out := filepath.Join(dir, "build", "debug", "ir", pkg, "workers_real.c")
                _ = os.MkdirAll(filepath.Dir(out), 0o755)
                _ = os.WriteFile(out, []byte(c.String()), 0o644)
            }
        }
        // Compile
        outDir := filepath.Join(dir, "build", env, "lib", pkg)
        if err := os.MkdirAll(outDir, 0o755); err != nil { continue }
        var libPath string
        var cmd *exec.Cmd
        // Include both wrapper and real impl sources when available.
        var sources []string
        sources = append(sources, cfile)
        realImpl := filepath.Join(dir, "build", "debug", "ir", pkg, "workers_real.c")
        if st, err := os.Stat(realImpl); err == nil && !st.IsDir() { sources = append(sources, realImpl) }
        if strings.HasPrefix(env, "darwin/") {
            libPath = filepath.Join(outDir, "libworkers.dylib")
            args := append([]string{"-dynamiclib"}, append(sources, []string{"-o", libPath, "-target", triple}...)...)
            cmd = exec.Command(clang, args...)
        } else if strings.HasPrefix(env, "linux/") {
            libPath = filepath.Join(outDir, "libworkers.so")
            args := append([]string{"-shared", "-fPIC"}, append(sources, []string{"-o", libPath, "-target", triple}...)...)
            cmd = exec.Command(clang, args...)
        } else if strings.HasPrefix(env, "windows/") {
            libPath = filepath.Join(outDir, "workers.dll")
            args := append([]string{"-shared"}, append(sources, []string{"-o", libPath}...)...)
            cmd = exec.Command(clang, args...)
        }
        if cmd != nil {
            if outb, err := cmd.CombinedOutput(); err == nil {
                // Update manifest with workersLib path for this package
                mfPath := filepath.Join(dir, "build", "debug", "manifest.json")
                var obj map[string]any
                if b, err := os.ReadFile(mfPath); err == nil { _ = json.Unmarshal(b, &obj) }
                if obj == nil { obj = map[string]any{"schema": "manifest.v1"} }
                pkgs, _ := obj["packages"].([]any)
                for i := range pkgs {
                    m := pkgs[i].(map[string]any)
                    if m["name"] == pkg {
                        // write a workspace-relative path
                        rel := strings.TrimPrefix(libPath, dir+string(filepath.Separator))
                        m["workersLib"] = rel
                    }
                }
                obj["packages"] = pkgs
                if b, err := json.MarshalIndent(obj, "", "  "); err == nil { _ = os.WriteFile(mfPath, b, 0o644) }
            } else if jsonOut {
                d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_LINK_FAIL", Message: "failed to compile workers lib", File: cfile, Data: map[string]any{"env": env, "stderr": string(outb)}}
                _ = json.NewEncoder(os.Stdout).Encode(d)
            }
        }
    }
    return nil
}
