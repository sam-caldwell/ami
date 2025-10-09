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
            type gb struct{ Family, Name, Source, Args string; N int; Grid [3]int; TPG [3]int }
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
                            argsSpec, _ := gm["args"].(string)
                            n := 0
                            if nn, ok := gm["n"].(float64); ok { n = int(nn) }
                            var grid [3]int
                            var tpg [3]int
                            if arr, ok := gm["grid"].([]any); ok {
                                for i := 0; i < len(arr) && i < 3; i++ { if v, ok := arr[i].(float64); ok { grid[i] = int(v) } }
                            }
                            if arr, ok := gm["tpg"].([]any); ok {
                                for i := 0; i < len(arr) && i < 3; i++ { if v, ok := arr[i].(float64); ok { tpg[i] = int(v) } }
                            }
                            list = append(list, gb{Family: strings.ToLower(fam), Name: kn, Source: src, Args: argsSpec, N: n, Grid: grid, TPG: tpg})
                        }
                        if len(list) > 0 {
                            gpuFuncs[fname] = list
                        }
                    }
                }
            }
            if len(metalWorkers) > 0 || len(gpuFuncs) > 0 {
                var c strings.Builder
                c.WriteString("#include <stdlib.h>\n#include <string.h>\n#include <stdint.h>\n")
                c.WriteString("#ifdef _WIN32\n#include <windows.h>\n#endif\n\n")
                c.WriteString("#if defined(__APPLE__) || defined(__linux__)\n#include <dlfcn.h>\nstatic void* _sym(const char* n){ return dlsym(RTLD_DEFAULT,n);}\n#endif\n")
                // Function pointer typedefs for runtime symbols
                c.WriteString("typedef unsigned char (*p_gpu_has)(long long);\n")
                c.WriteString("typedef void* (*p_metal_ctx_create)(void*);\n")
                c.WriteString("typedef void  (*p_metal_ctx_destroy)(void*);\n")
                c.WriteString("typedef void* (*p_metal_lib_compile)(void*);\n")
                c.WriteString("typedef void* (*p_metal_pipe_create)(void*, void*);\n")
                c.WriteString("typedef void* (*p_metal_alloc)(long long);\n")
                c.WriteString("typedef void  (*p_metal_free)(void*);\n")
                c.WriteString("typedef void  (*p_metal_copy_from_device)(void*, void*, long long);\n")
                c.WriteString("typedef void* (*p_metal_dispatch_1buf1u32)(void*, void*, void*, unsigned int, long long, long long, long long, long long, long long, long long);\n\n")
                // Minimal CUDA driver typedefs (opaque handles + essential entrypoints)
                c.WriteString("typedef int CUresult; typedef int CUdevice; typedef void* CUcontext; typedef void* CUmodule; typedef void* CUfunction; typedef unsigned long long CUdeviceptr;\n")
                c.WriteString("typedef CUresult (*p_cuInit)(unsigned int);\n")
                c.WriteString("typedef CUresult (*p_cuDeviceGet)(CUdevice*, int);\n")
                c.WriteString("typedef CUresult (*p_cuCtxCreate)(CUcontext*, unsigned int, CUdevice);\n")
                c.WriteString("typedef CUresult (*p_cuCtxDestroy)(CUcontext);\n")
                c.WriteString("typedef CUresult (*p_cuModuleLoadData)(CUmodule*, const void*);\n")
                c.WriteString("typedef CUresult (*p_cuModuleGetFunction)(CUfunction*, CUmodule, const char*);\n")
                c.WriteString("typedef CUresult (*p_cuMemAlloc)(CUdeviceptr*, size_t);\n")
                c.WriteString("typedef CUresult (*p_cuMemFree)(CUdeviceptr);\n")
                c.WriteString("typedef CUresult (*p_cuMemcpyDtoH)(void*, CUdeviceptr, size_t);\n")
                c.WriteString("typedef CUresult (*p_cuLaunchKernel)(CUfunction, unsigned int, unsigned int, unsigned int, unsigned int, unsigned int, unsigned int, unsigned int, void*, void**, void**);\n")
                c.WriteString("typedef CUresult (*p_cuCtxSynchronize)(void);\n")
                c.WriteString("typedef CUresult (*p_cuModuleUnload)(CUmodule);\n\n")

                // Minimal OpenCL typedefs (opaque pointers)
                c.WriteString("typedef void* cl_platform_id; typedef void* cl_device_id; typedef void* cl_context; typedef void* cl_command_queue; typedef void* cl_program; typedef void* cl_kernel; typedef void* cl_mem; typedef unsigned long cl_ulong; typedef unsigned int cl_uint; typedef long cl_int; typedef size_t cl_size_t;\n")
                c.WriteString("#define CL_DEVICE_TYPE_DEFAULT 1UL\n#define CL_MEM_READ_WRITE (1UL<<0)\n#define CL_TRUE 1\n#define CL_SUCCESS 0\n")
                c.WriteString("typedef cl_int (*p_clGetPlatformIDs)(cl_uint, cl_platform_id*, cl_uint*);\n")
                c.WriteString("typedef cl_int (*p_clGetDeviceIDs)(cl_platform_id, cl_ulong, cl_uint, cl_device_id*, cl_uint*);\n")
                c.WriteString("typedef cl_context (*p_clCreateContext)(const void*, cl_uint, const cl_device_id*, void*, void*, cl_int*);\n")
                c.WriteString("typedef cl_command_queue (*p_clCreateCommandQueue)(cl_context, cl_device_id, cl_ulong, cl_int*);\n")
                c.WriteString("typedef cl_program (*p_clCreateProgramWithSource)(cl_context, cl_uint, const char**, const size_t*, cl_int*);\n")
                c.WriteString("typedef cl_int (*p_clBuildProgram)(cl_program, cl_uint, const cl_device_id*, const char*, void*, void*);\n")
                c.WriteString("typedef cl_kernel (*p_clCreateKernel)(cl_program, const char*, cl_int*);\n")
                c.WriteString("typedef cl_mem (*p_clCreateBuffer)(cl_context, cl_ulong, size_t, void*, cl_int*);\n")
                c.WriteString("typedef cl_int (*p_clSetKernelArg)(cl_kernel, cl_uint, size_t, const void*);\n")
                c.WriteString("typedef cl_int (*p_clEnqueueNDRangeKernel)(cl_command_queue, cl_kernel, cl_uint, const size_t*, const size_t*, const size_t*, cl_uint, const void*, void*);\n")
                c.WriteString("typedef cl_int (*p_clEnqueueReadBuffer)(cl_command_queue, cl_mem, cl_uint, size_t, size_t, void*, cl_uint, const void*, void*);\n")
                c.WriteString("typedef cl_int (*p_clFinish)(cl_command_queue);\n")
                c.WriteString("typedef cl_int (*p_clReleaseMemObject)(cl_mem);\n")
                c.WriteString("typedef cl_int (*p_clReleaseKernel)(cl_kernel);\n")
                c.WriteString("typedef cl_int (*p_clReleaseProgram)(cl_program);\n")
                c.WriteString("typedef cl_int (*p_clReleaseCommandQueue)(cl_command_queue);\n")
                c.WriteString("typedef cl_int (*p_clReleaseContext)(cl_context);\n\n")
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
                    c.WriteString("    // metal backend\n")
                    // Select first metal block; embed its source and kernel name
                    c.WriteString("    const char* metal_src = NULL; const char* metal_kname = NULL; const char* metal_args = \"1buf1u32\"; unsigned int n = 8;\n")
                    c.WriteString("    long long gx=0, gy=1, gz=1, tx=1, ty=1, tz=1;\n")
                    // CUDA block (if present)
                    c.WriteString("    const char* cuda_ptx = NULL; const char* cuda_kname = NULL; const char* cuda_args = \"1buf1u32\";\n")
                    c.WriteString("    unsigned int cuda_gx=0, cuda_gy=1, cuda_gz=1, cuda_bx=1, cuda_by=1, cuda_bz=1;\n")
                    // OpenCL block (if present)
                    c.WriteString("    const char* ocl_src = NULL; const char* ocl_kname = NULL; const char* ocl_args = \"1buf1u32\";\n")
                    c.WriteString("    size_t ocl_gx=0, ocl_gy=1, ocl_gz=1, ocl_tx=1, ocl_ty=1, ocl_tz=1;\n")
                    for _, blk := range blocks {
                        if blk.Family == "metal" && blk.Source != "" {
                            esc := strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(blk.Source)
                            c.WriteString("    metal_src = \""); c.WriteString(esc); c.WriteString("\";\n")
                            if blk.Name != "" {
                                kn := strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(blk.Name)
                                c.WriteString("    metal_kname = \""); c.WriteString(kn); c.WriteString("\";\n")
                            }
                            if blk.N > 0 { c.WriteString("    n = "); c.WriteString(strconv.Itoa(blk.N)); c.WriteString(";\n") }
                            if blk.Args != "" { c.WriteString("    metal_args = \""); c.WriteString(strings.ReplaceAll(blk.Args, "\"", "\\\"")); c.WriteString("\";\n") }
                            // grid/tpg
                            if blk.Grid != [3]int{} {
                                gx := blk.Grid[0]; if gx <= 0 { gx = 1 }
                                gy := blk.Grid[1]; if gy <= 0 { gy = 1 }
                                gz := blk.Grid[2]; if gz <= 0 { gz = 1 }
                                c.WriteString("    gx = "); c.WriteString(strconv.Itoa(gx)); c.WriteString("; gy = "); c.WriteString(strconv.Itoa(gy)); c.WriteString("; gz = "); c.WriteString(strconv.Itoa(gz)); c.WriteString(";\n")
                            } else {
                                c.WriteString("    gx = (long long)n; gy = 1; gz = 1;\n")
                            }
                            if blk.TPG != [3]int{} {
                                tx := blk.TPG[0]; if tx <= 0 { tx = 1 }
                                ty := blk.TPG[1]; if ty <= 0 { ty = 1 }
                                tz := blk.TPG[2]; if tz <= 0 { tz = 1 }
                                c.WriteString("    tx = "); c.WriteString(strconv.Itoa(tx)); c.WriteString("; ty = "); c.WriteString(strconv.Itoa(ty)); c.WriteString("; tz = "); c.WriteString(strconv.Itoa(tz)); c.WriteString(";\n")
                            }
                            break
                        }
                        if blk.Family == "opencl" && blk.Source != "" {
                            esc2 := strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(blk.Source)
                            c.WriteString("    ocl_src = \""); c.WriteString(esc2); c.WriteString("\";\n")
                            if blk.Name != "" {
                                kn2 := strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(blk.Name)
                                c.WriteString("    ocl_kname = \""); c.WriteString(kn2); c.WriteString("\";\n")
                            }
                            // adopt same n/grid defaults
                            if blk.N > 0 { c.WriteString("    n = "); c.WriteString(strconv.Itoa(blk.N)); c.WriteString(";\n") }
                            if blk.Args != "" { c.WriteString("    ocl_args = \""); c.WriteString(strings.ReplaceAll(blk.Args, "\"", "\\\"")); c.WriteString("\";\n") }
                            if blk.Grid != [3]int{} {
                                gx := blk.Grid[0]; if gx <= 0 { gx = 1 }
                                gy := blk.Grid[1]; if gy <= 0 { gy = 1 }
                                gz := blk.Grid[2]; if gz <= 0 { gz = 1 }
                                c.WriteString("    ocl_gx = "); c.WriteString(strconv.Itoa(gx)); c.WriteString("; ocl_gy = "); c.WriteString(strconv.Itoa(gy)); c.WriteString("; ocl_gz = "); c.WriteString(strconv.Itoa(gz)); c.WriteString(";\n")
                            } else {
                                c.WriteString("    ocl_gx = (size_t)n; ocl_gy = 1; ocl_gz = 1;\n")
                            }
                            if blk.TPG != [3]int{} {
                                tx := blk.TPG[0]; if tx <= 0 { tx = 1 }
                                ty := blk.TPG[1]; if ty <= 0 { ty = 1 }
                                tz := blk.TPG[2]; if tz <= 0 { tz = 1 }
                                c.WriteString("    ocl_tx = "); c.WriteString(strconv.Itoa(tx)); c.WriteString("; ocl_ty = "); c.WriteString(strconv.Itoa(ty)); c.WriteString("; ocl_tz = "); c.WriteString(strconv.Itoa(tz)); c.WriteString(";\n")
                            }
                        }
                        if blk.Family == "cuda" && blk.Source != "" {
                            // Embed PTX for CUDA and capture kernel name and dims
                            esc3 := strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(blk.Source)
                            c.WriteString("    cuda_ptx = \""); c.WriteString(esc3); c.WriteString("\";\n")
                            if blk.Name != "" {
                                kn3 := strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(blk.Name)
                                c.WriteString("    cuda_kname = \""); c.WriteString(kn3); c.WriteString("\";\n")
                            }
                            if blk.N > 0 { c.WriteString("    n = "); c.WriteString(strconv.Itoa(blk.N)); c.WriteString(";\n") }
                            if blk.Args != "" { c.WriteString("    cuda_args = \""); c.WriteString(strings.ReplaceAll(blk.Args, "\"", "\\\"")); c.WriteString("\";\n") }
                            if blk.Grid != [3]int{} {
                                gx := blk.Grid[0]; if gx <= 0 { gx = 1 }
                                gy := blk.Grid[1]; if gy <= 0 { gy = 1 }
                                gz := blk.Grid[2]; if gz <= 0 { gz = 1 }
                                c.WriteString("    cuda_gx = "); c.WriteString(strconv.Itoa(gx)); c.WriteString("; cuda_gy = "); c.WriteString(strconv.Itoa(gy)); c.WriteString("; cuda_gz = "); c.WriteString(strconv.Itoa(gz)); c.WriteString(";\n")
                            } else {
                                c.WriteString("    cuda_gx = n; cuda_gy = 1; cuda_gz = 1;\n")
                            }
                            if blk.TPG != [3]int{} {
                                bx := blk.TPG[0]; if bx <= 0 { bx = 1 }
                                by := blk.TPG[1]; if by <= 0 { by = 1 }
                                bz := blk.TPG[2]; if bz <= 0 { bz = 1 }
                                c.WriteString("    cuda_bx = "); c.WriteString(strconv.Itoa(bx)); c.WriteString("; cuda_by = "); c.WriteString(strconv.Itoa(by)); c.WriteString("; cuda_bz = "); c.WriteString(strconv.Itoa(bz)); c.WriteString(";\n")
                            }
                        }
                    }
                    c.WriteString("    if (metal_src) {\n")
                    c.WriteString("      p_gpu_has f_has = (p_gpu_has)_sym(\"ami_rt_gpu_has\");\n")
                    c.WriteString("      if (f_has && f_has(0)) {\n")
                    c.WriteString("        p_metal_ctx_create f_ctx_create = (p_metal_ctx_create)_sym(\"ami_rt_metal_ctx_create\");\n")
                    c.WriteString("        p_metal_ctx_destroy f_ctx_destroy = (p_metal_ctx_destroy)_sym(\"ami_rt_metal_ctx_destroy\");\n")
                    c.WriteString("        p_metal_lib_compile f_lib_compile = (p_metal_lib_compile)_sym(\"ami_rt_metal_lib_compile\");\n")
                    c.WriteString("        p_metal_pipe_create f_pipe_create = (p_metal_pipe_create)_sym(\"ami_rt_metal_pipe_create\");\n")
                    c.WriteString("        p_metal_alloc f_alloc = (p_metal_alloc)_sym(\"ami_rt_metal_alloc\");\n")
                    c.WriteString("        p_metal_free f_free = (p_metal_free)_sym(\"ami_rt_metal_free\");\n")
                    c.WriteString("        p_metal_copy_from_device f_copy_from = (p_metal_copy_from_device)_sym(\"ami_rt_metal_copy_from_device\");\n")
                    c.WriteString("        p_metal_dispatch_1buf1u32 f_dispatch = (p_metal_dispatch_1buf1u32)_sym(\"ami_rt_metal_dispatch_blocking_1buf1u32\");\n")
                    c.WriteString("        if (!f_ctx_create||!f_ctx_destroy||!f_lib_compile||!f_pipe_create||!f_alloc||!f_free||!f_copy_from||!f_dispatch){ if (err) *err = strdup(\"metal symbols\"); return NULL; }\n")
                    c.WriteString("        void* ctx = f_ctx_create(NULL); if (!ctx) { if (err) *err = strdup(\"metal ctx\"); return NULL; }\n")
                    c.WriteString("        void* lib = f_lib_compile((void*)metal_src); if (!lib) { if (err) *err = strdup(\"metal lib\"); return NULL; }\n")
                    c.WriteString("        const char* kname = metal_kname ? metal_kname : \"main\"; void* pipe = f_pipe_create(lib, (void*)kname); if (!pipe) { if (err) *err = strdup(\"metal pipe\"); return NULL; }\n")
                    c.WriteString("        void* dbuf = f_alloc((long long)n * 8); if (!dbuf) { if (err) *err = strdup(\"metal alloc\"); return NULL; }\n")
                    c.WriteString("        (void)f_dispatch(ctx, pipe, dbuf, n, gx, gy, gz, tx, ty, tz);\n")
                    c.WriteString("        long* raw = (long*)malloc((size_t)n * 8); if (!raw) { if (err) *err = strdup(\"oom\"); return NULL; }\n")
                    c.WriteString("        f_copy_from((void*)raw, dbuf, (long long)n * 8);\n")
                    c.WriteString("      size_t cap = (size_t)n * 24 + 2; char* js = (char*)malloc(cap); if (!js) { if (err) *err = strdup(\"oom\"); return NULL; }\n")
                    c.WriteString("      size_t pos = 0; js[pos++]='['; for (unsigned int i=0;i<n;i++){ if (i>0) js[pos++]=','; pos += (size_t)snprintf(js+pos, cap-pos, \"%ld\", raw[i]); } js[pos++]=']'; *out_len=(int)pos;\n")
                    c.WriteString("      free(raw); f_free(dbuf); f_ctx_destroy(ctx); return (const char*)js;\n")
                    c.WriteString("      }\n")
                    c.WriteString("    }\n")
                    c.WriteString("#endif\n")
                    // Minimal CUDA branch: dlopen libcuda and cuInit(0); return JSON diag on success/failure
                    c.WriteString("#if defined(__APPLE__) || defined(__linux__) || defined(_WIN32)\n")
                    c.WriteString("    if (ami_rt_gpu_has(1)) {\n")
                    c.WriteString("      void* libc = NULL;\n")
                    c.WriteString("#if defined(__APPLE__)\n      libc = dlopen(\"libcuda.dylib\", RTLD_LAZY);\n#else\n#ifdef _WIN32\n      libc = (void*)LoadLibraryA(\"nvcuda.dll\");\n#else\n      libc = dlopen(\"libcuda.so.1\", RTLD_LAZY); if(!libc) libc = dlopen(\"libcuda.so\", RTLD_LAZY);\n#endif\n#endif\n")
                    c.WriteString("      if (!libc) { if (err) *err = strdup(\"cuda dlopen\"); return NULL; }\n")
                    c.WriteString("#if defined(_WIN32)\n      p_cuInit cuInit = (p_cuInit)GetProcAddress((HMODULE)libc, \"cuInit\");\n#else\n      p_cuInit cuInit = (p_cuInit)dlsym(libc, \"cuInit\");\n#endif\n")
                    c.WriteString("      if (!cuInit) { if (err) *err = strdup(\"cuda symbol\"); return NULL; }\n")
                    c.WriteString("      int rc = cuInit(0);\n")
                    c.WriteString("      const char* ok = \"{\\\"backend\\\":\\\"cuda\\\",\\\"status\\\":\\\"ok\\\"}\";\n")
                    c.WriteString("      const char* bad = \"{\\\"backend\\\":\\\"cuda\\\",\\\"status\\\":\\\"init_fail\\\"}\";\n")
                    c.WriteString("      const char* ms = (rc==0)? ok : bad; size_t L = strlen(ms); char* js = (char*)malloc(L); if(!js){ if(err)*err=strdup(\"oom\"); return NULL; } memcpy(js, ms, L); *out_len=(int)L; return (const char*)js;\n")
                    c.WriteString("    }\n")
                    c.WriteString("#endif\n")
                    // CUDA branch: dlopen libcuda + driver API launch
                    c.WriteString("#if defined(__APPLE__) || defined(__linux__) || defined(_WIN32)\n")
                    c.WriteString("    if (cuda_ptx && ami_rt_gpu_has(1)) {\n")
                    // Load driver library
                    c.WriteString("      void* libc = NULL;\n")
                    c.WriteString("#if defined(__APPLE__)\n      libc = dlopen(\"libcuda.dylib\", RTLD_LAZY);\n#elif defined(_WIN32)\n      libc = (void*)LoadLibraryA(\"nvcuda.dll\");\n#else\n      libc = dlopen(\"libcuda.so.1\", RTLD_LAZY); if(!libc) libc = dlopen(\"libcuda.so\", RTLD_LAZY);\n#endif\n")
                    c.WriteString("      if (!libc) { if (err) *err = strdup(\"cuda dlopen\"); return NULL; }\n")
                    // Resolve symbols
                    c.WriteString("#if defined(_WIN32)\n      p_cuInit cuInit = (p_cuInit)GetProcAddress((HMODULE)libc, \"cuInit\");\n      p_cuDeviceGet cuDeviceGet = (p_cuDeviceGet)GetProcAddress((HMODULE)libc, \"cuDeviceGet\");\n      p_cuCtxCreate cuCtxCreate = (p_cuCtxCreate)GetProcAddress((HMODULE)libc, \"cuCtxCreate_v2\");\n      p_cuCtxDestroy cuCtxDestroy = (p_cuCtxDestroy)GetProcAddress((HMODULE)libc, \"cuCtxDestroy_v2\");\n      p_cuModuleLoadData cuModuleLoadData = (p_cuModuleLoadData)GetProcAddress((HMODULE)libc, \"cuModuleLoadData\");\n      p_cuModuleGetFunction cuModuleGetFunction = (p_cuModuleGetFunction)GetProcAddress((HMODULE)libc, \"cuModuleGetFunction\");\n      p_cuMemAlloc cuMemAlloc = (p_cuMemAlloc)GetProcAddress((HMODULE)libc, \"cuMemAlloc_v2\");\n      p_cuMemFree cuMemFree = (p_cuMemFree)GetProcAddress((HMODULE)libc, \"cuMemFree_v2\");\n      p_cuMemcpyDtoH cuMemcpyDtoH = (p_cuMemcpyDtoH)GetProcAddress((HMODULE)libc, \"cuMemcpyDtoH_v2\");\n      p_cuLaunchKernel cuLaunchKernel = (p_cuLaunchKernel)GetProcAddress((HMODULE)libc, \"cuLaunchKernel\");\n      p_cuCtxSynchronize cuCtxSynchronize = (p_cuCtxSynchronize)GetProcAddress((HMODULE)libc, \"cuCtxSynchronize\");\n      p_cuModuleUnload cuModuleUnload = (p_cuModuleUnload)GetProcAddress((HMODULE)libc, \"cuModuleUnload\");\n#else\n      p_cuInit cuInit = (p_cuInit)dlsym(libc, \"cuInit\");\n      p_cuDeviceGet cuDeviceGet = (p_cuDeviceGet)dlsym(libc, \"cuDeviceGet\");\n      p_cuCtxCreate cuCtxCreate = (p_cuCtxCreate)dlsym(libc, \"cuCtxCreate_v2\");\n      p_cuCtxDestroy cuCtxDestroy = (p_cuCtxDestroy)dlsym(libc, \"cuCtxDestroy_v2\");\n      p_cuModuleLoadData cuModuleLoadData = (p_cuModuleLoadData)dlsym(libc, \"cuModuleLoadData\");\n      p_cuModuleGetFunction cuModuleGetFunction = (p_cuModuleGetFunction)dlsym(libc, \"cuModuleGetFunction\");\n      p_cuMemAlloc cuMemAlloc = (p_cuMemAlloc)dlsym(libc, \"cuMemAlloc_v2\");\n      p_cuMemFree cuMemFree = (p_cuMemFree)dlsym(libc, \"cuMemFree_v2\");\n      p_cuMemcpyDtoH cuMemcpyDtoH = (p_cuMemcpyDtoH)dlsym(libc, \"cuMemcpyDtoH_v2\");\n      p_cuLaunchKernel cuLaunchKernel = (p_cuLaunchKernel)dlsym(libc, \"cuLaunchKernel\");\n      p_cuCtxSynchronize cuCtxSynchronize = (p_cuCtxSynchronize)dlsym(libc, \"cuCtxSynchronize\");\n      p_cuModuleUnload cuModuleUnload = (p_cuModuleUnload)dlsym(libc, \"cuModuleUnload\");\n#endif\n")
                    c.WriteString("      if(!cuInit||!cuDeviceGet||!cuCtxCreate||!cuCtxDestroy||!cuModuleLoadData||!cuModuleGetFunction||!cuMemAlloc||!cuMemFree||!cuMemcpyDtoH||!cuLaunchKernel){ if (err) *err = strdup(\"cuda symbols\"); return NULL; }\n")
                    c.WriteString("      if (cuInit(0) != 0) { if (err) *err = strdup(\"cuda init\"); return NULL; }\n")
                    c.WriteString("      CUdevice dev = 0; if (cuDeviceGet(&dev, 0) != 0) { if (err) *err = strdup(\"cuda device\"); return NULL; }\n")
                    c.WriteString("      CUcontext cuctx = 0; if (cuCtxCreate(&cuctx, 0, dev) != 0 || !cuctx) { if (err) *err = strdup(\"cuda ctx\"); return NULL; }\n")
                    c.WriteString("      CUmodule mod = 0; if (cuModuleLoadData(&mod, (const void*)cuda_ptx) != 0 || !mod) { if (err) *err = strdup(\"cuda module\"); return NULL; }\n")
                    c.WriteString("      const char* kname3 = cuda_kname ? cuda_kname : \"main\"; CUfunction fn = 0; if (cuModuleGetFunction(&fn, mod, kname3) != 0 || !fn) { if (err) *err = strdup(\"cuda kernel\"); return NULL; }\n")
                    c.WriteString("      CUdeviceptr dptr = 0; size_t bytes = (size_t)n * 8; if (cuMemAlloc(&dptr, bytes) != 0 || !dptr) { if (err) *err = strdup(\"cuda alloc\"); return NULL; }\n")
                    c.WriteString("      unsigned int n32 = n; void* args2[2]; void* args1[1]; void** argv = args2; args2[0] = &dptr; args2[1] = &n32; args1[0] = &dptr;\n")
                    c.WriteString("      if (cuda_args && cuda_args[0]=='1' && cuda_args[1]=='b' && cuda_args[2]=='u' && cuda_args[3]=='f') { argv = args1; }\n")
                    c.WriteString("      unsigned int gx3 = cuda_gx ? cuda_gx : n; unsigned int gy3 = cuda_gy ? cuda_gy : 1; unsigned int gz3 = cuda_gz ? cuda_gz : 1;\n")
                    c.WriteString("      unsigned int bx3 = cuda_bx ? cuda_bx : 1; unsigned int by3 = cuda_by ? cuda_by : 1; unsigned int bz3 = cuda_bz ? cuda_bz : 1;\n")
                    c.WriteString("      if (cuLaunchKernel(fn, gx3, gy3, gz3, bx3, by3, bz3, 0, (void*)0, argv, (void**)0) != 0) { if (err) *err = strdup(\"cuda launch\"); return NULL; }\n")
                    c.WriteString("      if (cuCtxSynchronize) { (void)cuCtxSynchronize(); }\n")
                    c.WriteString("      long* raw = (long*)malloc(bytes); if (!raw) { if (err) *err = strdup(\"oom\"); return NULL; }\n")
                    c.WriteString("      if (cuMemcpyDtoH((void*)raw, dptr, bytes) != 0) { if (err) *err = strdup(\"cuda memcpy\"); return NULL; }\n")
                    c.WriteString("      size_t cap = (size_t)n * 24 + 2; char* js = (char*)malloc(cap); if (!js) { if (err) *err = strdup(\"oom\"); (void)cuMemFree(dptr); if (cuModuleUnload) (void)cuModuleUnload(mod); (void)cuCtxDestroy(cuctx); return NULL; }\n")
                    c.WriteString("      size_t pos = 0; js[pos++]='['; for (unsigned int i=0;i<n;i++){ if (i>0) js[pos++]=','; pos += (size_t)snprintf(js+pos, cap-pos, \"%ld\", raw[i]); } js[pos++]=']'; *out_len=(int)pos;\n")
                    c.WriteString("      free(raw); (void)cuMemFree(dptr); if (cuModuleUnload) (void)cuModuleUnload(mod); (void)cuCtxDestroy(cuctx); return (const char*)js;\n")
                    c.WriteString("    }\n")
                    c.WriteString("#endif\n")

                    // OpenCL branch: dynamic dispatch using dlsym
                    c.WriteString("#if defined(__APPLE__) || defined(__linux__)\n")
                    c.WriteString("    if (ocl_src && ami_rt_gpu_has(2)) {\n")
                    c.WriteString("      void* libcl = NULL;\n")
                    c.WriteString("#if defined(__APPLE__)\n      libcl = dlopen(\"/System/Library/Frameworks/OpenCL.framework/OpenCL\", RTLD_LAZY); if(!libcl) libcl = dlopen(\"libOpenCL.dylib\", RTLD_LAZY);\n#else\n      libcl = dlopen(\"libOpenCL.so.1\", RTLD_LAZY); if(!libcl) libcl = dlopen(\"libOpenCL.so\", RTLD_LAZY);\n#endif\n")
                    c.WriteString("      if (!libcl) { if (err) *err = strdup(\"opencl dlopen\"); return NULL; }\n")
                    c.WriteString("      // resolve required symbols\n")
                    c.WriteString("      p_clGetPlatformIDs clGetPlatformIDs = (p_clGetPlatformIDs)dlsym(libcl, \"clGetPlatformIDs\");\n")
                    c.WriteString("      p_clGetDeviceIDs clGetDeviceIDs = (p_clGetDeviceIDs)dlsym(libcl, \"clGetDeviceIDs\");\n")
                    c.WriteString("      p_clCreateContext clCreateContext = (p_clCreateContext)dlsym(libcl, \"clCreateContext\");\n")
                    c.WriteString("      p_clCreateCommandQueue clCreateCommandQueue = (p_clCreateCommandQueue)dlsym(libcl, \"clCreateCommandQueue\");\n")
                    c.WriteString("      p_clCreateProgramWithSource clCreateProgramWithSource = (p_clCreateProgramWithSource)dlsym(libcl, \"clCreateProgramWithSource\");\n")
                    c.WriteString("      p_clBuildProgram clBuildProgram = (p_clBuildProgram)dlsym(libcl, \"clBuildProgram\");\n")
                    c.WriteString("      p_clCreateKernel clCreateKernel = (p_clCreateKernel)dlsym(libcl, \"clCreateKernel\");\n")
                    c.WriteString("      p_clCreateBuffer clCreateBuffer = (p_clCreateBuffer)dlsym(libcl, \"clCreateBuffer\");\n")
                    c.WriteString("      p_clSetKernelArg clSetKernelArg = (p_clSetKernelArg)dlsym(libcl, \"clSetKernelArg\");\n")
                    c.WriteString("      p_clEnqueueNDRangeKernel clEnqueueNDRangeKernel = (p_clEnqueueNDRangeKernel)dlsym(libcl, \"clEnqueueNDRangeKernel\");\n")
                    c.WriteString("      p_clEnqueueReadBuffer clEnqueueReadBuffer = (p_clEnqueueReadBuffer)dlsym(libcl, \"clEnqueueReadBuffer\");\n")
                    c.WriteString("      p_clFinish clFinish = (p_clFinish)dlsym(libcl, \"clFinish\");\n")
                    c.WriteString("      p_clReleaseMemObject clReleaseMemObject = (p_clReleaseMemObject)dlsym(libcl, \"clReleaseMemObject\");\n")
                    c.WriteString("      p_clReleaseKernel clReleaseKernel = (p_clReleaseKernel)dlsym(libcl, \"clReleaseKernel\");\n")
                    c.WriteString("      p_clReleaseProgram clReleaseProgram = (p_clReleaseProgram)dlsym(libcl, \"clReleaseProgram\");\n")
                    c.WriteString("      p_clReleaseCommandQueue clReleaseCommandQueue = (p_clReleaseCommandQueue)dlsym(libcl, \"clReleaseCommandQueue\");\n")
                    c.WriteString("      p_clReleaseContext clReleaseContext = (p_clReleaseContext)dlsym(libcl, \"clReleaseContext\");\n")
                    c.WriteString("      if(!clGetPlatformIDs||!clGetDeviceIDs||!clCreateContext||!clCreateCommandQueue||!clCreateProgramWithSource||!clBuildProgram||!clCreateKernel||!clCreateBuffer||!clSetKernelArg||!clEnqueueNDRangeKernel||!clEnqueueReadBuffer||!clFinish||!clReleaseMemObject||!clReleaseKernel||!clReleaseProgram||!clReleaseCommandQueue||!clReleaseContext){ if (err) *err = strdup(\"opencl symbols\"); return NULL; }\n")
                    c.WriteString("      cl_platform_id platform = 0; cl_uint nplat = 0; if (clGetPlatformIDs(1, &platform, &nplat) != CL_SUCCESS || nplat == 0) { if (err) *err = strdup(\"opencl platform\"); return NULL; }\n")
                    c.WriteString("      cl_device_id device = 0; cl_uint ndev = 0; if (clGetDeviceIDs(platform, CL_DEVICE_TYPE_DEFAULT, 1, &device, &ndev) != CL_SUCCESS || ndev == 0) { if (err) *err = strdup(\"opencl device\"); return NULL; }\n")
                    c.WriteString("      cl_int e = 0; cl_context ctx = clCreateContext(NULL, 1, &device, NULL, NULL, &e); if (!ctx || e != CL_SUCCESS) { if (err) *err = strdup(\"opencl ctx\"); return NULL; }\n")
                    c.WriteString("      cl_command_queue q = clCreateCommandQueue(ctx, device, 0, &e); if (!q || e != CL_SUCCESS) { if (err) *err = strdup(\"opencl queue\"); return NULL; }\n")
                    c.WriteString("      const char* srcs[1] = { ocl_src }; cl_program prog = clCreateProgramWithSource(ctx, 1, srcs, NULL, &e); if (!prog || e != CL_SUCCESS) { if (err) *err = strdup(\"opencl program\"); return NULL; }\n")
                    c.WriteString("      if (clBuildProgram(prog, 1, &device, NULL, NULL, NULL) != CL_SUCCESS) { if (err) *err = strdup(\"opencl build\"); return NULL; }\n")
                    c.WriteString("      const char* kname2 = ocl_kname ? ocl_kname : \"main\"; cl_kernel kern = clCreateKernel(prog, kname2, &e); if (!kern || e != CL_SUCCESS) { if (err) *err = strdup(\"opencl kernel\"); return NULL; }\n")
                    c.WriteString("      cl_mem buf = clCreateBuffer(ctx, CL_MEM_READ_WRITE, (size_t)n * 8, NULL, &e); if (!buf || e != CL_SUCCESS) { if (err) *err = strdup(\"opencl buffer\"); return NULL; }\n")
                    c.WriteString("      // set args: out buffer (arg0) and optionally n (arg1) depending on args spec\n")
                    c.WriteString("      unsigned int ncopy = n; if (clSetKernelArg(kern, 0, sizeof(cl_mem), &buf) != CL_SUCCESS) { if (err) *err = strdup(\"opencl args buf\"); return NULL; }\n")
                    c.WriteString("      if (!ocl_args || (ocl_args && ocl_args[0]=='1' && ocl_args[1]=='b' && ocl_args[2]=='u' && ocl_args[3]=='f' && ocl_args[4]=='1')) { if (clSetKernelArg(kern, 1, sizeof(unsigned int), &ncopy) != CL_SUCCESS) { if (err) *err = strdup(\"opencl args n\"); return NULL; } }\n")
                    c.WriteString("      size_t gws3[3] = { ocl_gx ? ocl_gx : (size_t)n, ocl_gy ? ocl_gy : 1, ocl_gz ? ocl_gz : 1 };\n")
                    c.WriteString("      size_t lws3[3] = { ocl_tx ? ocl_tx : 1, ocl_ty ? ocl_ty : 1, ocl_tz ? ocl_tz : 1 };\n")
                    c.WriteString("      cl_uint work_dim = 1; if (gws3[2] > 1 || lws3[2] > 1) work_dim = 3; else if (gws3[1] > 1 || lws3[1] > 1) work_dim = 2;\n")
                    c.WriteString("      if (clEnqueueNDRangeKernel(q, kern, work_dim, NULL, gws3, lws3, 0, NULL, NULL) != CL_SUCCESS) { if (err) *err = strdup(\"opencl enqueue\"); return NULL; }\n")
                    c.WriteString("      if (clFinish(q) != CL_SUCCESS) { if (err) *err = strdup(\"opencl finish\"); return NULL; }\n")
                    c.WriteString("      long* raw = (long*)malloc((size_t)n * 8); if (!raw) { if (err) *err = strdup(\"oom\"); return NULL; }\n")
                    c.WriteString("      if (clEnqueueReadBuffer(q, buf, CL_TRUE, 0, (size_t)n * 8, raw, 0, NULL, NULL) != CL_SUCCESS) { if (err) *err = strdup(\"opencl read\"); return NULL; }\n")
                    c.WriteString("      size_t cap = (size_t)n * 24 + 2; char* js = (char*)malloc(cap); if (!js) { if (err) *err = strdup(\"oom\"); return NULL; }\n")
                    c.WriteString("      size_t pos = 0; js[pos++]='['; for (unsigned int i=0;i<n;i++){ if (i>0) js[pos++]=','; pos += (size_t)snprintf(js+pos, cap-pos, \"%ld\", raw[i]); } js[pos++]=']'; *out_len=(int)pos;\n")
                    c.WriteString("      free(raw); clReleaseMemObject(buf); clReleaseKernel(kern); clReleaseProgram(prog); clReleaseCommandQueue(q); clReleaseContext(ctx); return (const char*)js;\n")
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
