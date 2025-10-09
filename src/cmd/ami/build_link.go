package main

import (
    "encoding/json"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"

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
        if lg := getRootLogger(); lg != nil {
            lg.Info("build.toolchain.missing", map[string]any{"tool": "clang"})
            lg.Info("toolchain.missing", map[string]any{"tool": "clang"})
        }
        return
    }
    if lg := getRootLogger(); lg != nil {
        if ver, verr := be.ToolVersion(clang); verr == nil {
            lg.Info("toolchain.clang", map[string]any{"path": clang, "version": ver})
        } else {
            lg.Info("toolchain.clang", map[string]any{"path": clang})
        }
    }
	// Resolve binary name
	binName := "app"
	if mp := ws.FindPackage("main"); mp != nil && mp.Name != "" {
		binName = mp.Name
	}
	// Per-env link
	for _, env := range envs {
		if containsEnv(buildNoLinkEnvs, env) {
			continue
		}
		// collect per-env objects
		var objs []string
		for _, e := range ws.Packages {
			glob := filepath.Join(dir, "build", env, "obj", e.Package.Name, "*.o")
			if matches, _ := filepath.Glob(glob); len(matches) > 0 {
				objs = append(objs, matches...)
			}
		}
		if len(objs) == 0 {
			continue
		}
		triple := be.TripleForEnv(env)
		rtDir := filepath.Join(dir, "build", "runtime", env)
		rtObj := filepath.Join(rtDir, "runtime.o")
		if _, stErr := os.Stat(rtObj); stErr != nil {
			if llPath, werr := be.WriteRuntimeLL(rtDir, triple, false); werr == nil {
				if cerr := be.CompileLLToObject(clang, llPath, rtObj, triple); cerr != nil {
					if lg := getRootLogger(); lg != nil {
						lg.Info("build.runtime.obj.fail", map[string]any{"error": cerr.Error(), "env": env})
					}
					if jsonOut {
						d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_OBJ_COMPILE_FAIL", Message: "failed to compile LLVM to object (runtime)", File: llPath, Data: map[string]any{"env": env, "what": "runtime"}}
						if te, ok := cerr.(llvme.ToolError); ok {
							if d.Data == nil {
								d.Data = map[string]any{}
							}
							d.Data["stderr"] = te.Stderr
						}
						_ = json.NewEncoder(out).Encode(d)
					}
				}
			} else {
				if lg := getRootLogger(); lg != nil {
					lg.Info("build.runtime.ll.fail", map[string]any{"error": werr.Error(), "env": env})
				}
				if jsonOut {
					d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_LLVM_EMIT", Message: werr.Error(), File: filepath.Join(rtDir, "runtime.ll"), Data: map[string]any{"env": env}}
					_ = json.NewEncoder(out).Encode(d)
				}
			}
		}
		// Darwin Metal shim: compile Objective-C runtime integration if on darwin
		var metalObj string
        if strings.HasPrefix(env, "darwin/") {
            metalSrc := filepath.Join(rtDir, "metal_shim.m")
            _ = os.WriteFile(metalSrc, []byte(`
#import <Foundation/Foundation.h>
#import <Metal/Metal.h>
#import <CoreFoundation/CoreFoundation.h>
#include <string.h>
// Extern C linkage for LLVM-visible runtime symbols
#ifdef __cplusplus
extern "C" {
#endif

typedef struct { CFTypeRef dev; CFTypeRef q; } AmiRtMetalCtx;
typedef struct { CFTypeRef lib; CFTypeRef dev; } AmiRtMetalLib;
typedef struct { CFTypeRef ps;  CFTypeRef dev; } AmiRtMetalPipe;
typedef struct { CFTypeRef buf; unsigned long len; CFTypeRef dev; } AmiRtMetalBuf;

unsigned char ami_rt_metal_available(void) {
    id<MTLDevice> dev = MTLCreateSystemDefaultDevice();
    return dev != nil ? 1 : 0;
}

void* ami_rt_metal_devices(void) { return NULL; }

void* ami_rt_metal_ctx_create(void* dev_in) {
    (void)dev_in; // future: honor device selection
    id<MTLDevice> dev = MTLCreateSystemDefaultDevice();
    if (!dev) return NULL;
    id<MTLCommandQueue> q = [dev newCommandQueue];
    if (!q) return NULL;
    AmiRtMetalCtx* ctx = (AmiRtMetalCtx*)malloc(sizeof(AmiRtMetalCtx));
    if (!ctx) return NULL;
    ctx->dev = CFBridgingRetain(dev);
    ctx->q   = CFBridgingRetain(q);
    return ctx;
}

void  ami_rt_metal_ctx_destroy(void* ctxp) {
    if (!ctxp) return;
    AmiRtMetalCtx* ctx = (AmiRtMetalCtx*)ctxp;
    if (ctx->q)   { CFRelease(ctx->q); }
    if (ctx->dev) { CFRelease(ctx->dev); }
    free(ctx);
}

void* ami_rt_metal_lib_compile(void* src) {
    id<MTLDevice> dev = MTLCreateSystemDefaultDevice();
    if (!dev) return NULL;
    const char* csrc = (const char*)src;
    NSString* code = [NSString stringWithUTF8String:csrc ? csrc : ""];
    NSError* e = nil;
    id<MTLLibrary> lib = [dev newLibraryWithSource:code options:nil error:&e];
    if (!lib) { return NULL; }
    AmiRtMetalLib* L = (AmiRtMetalLib*)malloc(sizeof(AmiRtMetalLib));
    if (!L) return NULL;
    L->dev = CFBridgingRetain(dev);
    L->lib = CFBridgingRetain(lib);
    return L;
}

void* ami_rt_metal_pipe_create(void* libp, void* namep) {
    if (!libp) return NULL;
    AmiRtMetalLib* L = (AmiRtMetalLib*)libp;
    id<MTLDevice> dev = (__bridge id<MTLDevice>)L->dev;
    id<MTLLibrary> lib = (__bridge id<MTLLibrary>)L->lib;
    const char* cname = (const char*)namep;
    NSString* fname = [NSString stringWithUTF8String:cname ? cname : ""];
    id<MTLFunction> fn = [lib newFunctionWithName:fname];
    if (!fn) { return NULL; }
    NSError* e = nil;
    id<MTLComputePipelineState> ps = [dev newComputePipelineStateWithFunction:fn error:&e];
    if (!ps) { return NULL; }
    AmiRtMetalPipe* P = (AmiRtMetalPipe*)malloc(sizeof(AmiRtMetalPipe));
    if (!P) return NULL;
    P->dev = CFRetain(L->dev);
    P->ps  = CFBridgingRetain(ps);
    return P;
}

void* ami_rt_metal_alloc(long long n) {
    id<MTLDevice> dev = MTLCreateSystemDefaultDevice();
    if (!dev) return NULL;
    id<MTLBuffer> b = [dev newBufferWithLength:(NSUInteger)n options:MTLResourceStorageModeShared];
    if (!b) return NULL;
    AmiRtMetalBuf* B = (AmiRtMetalBuf*)malloc(sizeof(AmiRtMetalBuf));
    if (!B) return NULL;
    B->dev = CFBridgingRetain(dev);
    B->buf = CFBridgingRetain(b);
    B->len = (unsigned long)n;
    return B;
}

void  ami_rt_metal_free(void* bufp) {
    if (!bufp) return;
    AmiRtMetalBuf* B = (AmiRtMetalBuf*)bufp;
    if (B->buf) { CFRelease(B->buf); }
    if (B->dev) { CFRelease(B->dev); }
    free(B);
}

void  ami_rt_metal_copy_to_device(void* dstp, void* src, long long n) {
    if (!dstp || !src || n <= 0) return;
    AmiRtMetalBuf* B = (AmiRtMetalBuf*)dstp;
    id<MTLBuffer> b = (__bridge id<MTLBuffer>)B->buf;
    if (!b) return;
    void* dst = [b contents];
    if ((unsigned long)n > B->len) return;
    memcpy(dst, src, (size_t)n);
}

void  ami_rt_metal_copy_from_device(void* dst, void* srcp, long long n) {
    if (!dst || !srcp || n <= 0) return;
    AmiRtMetalBuf* B = (AmiRtMetalBuf*)srcp;
    id<MTLBuffer> b = (__bridge id<MTLBuffer>)B->buf;
    if (!b) return;
    void* src = [b contents];
    if ((unsigned long)n > B->len) return;
    memcpy(dst, src, (size_t)n);
}

void* ami_rt_metal_dispatch_blocking(void* ctxp, void* pipep,
                                     long long gx, long long gy, long long gz,
                                     long long tx, long long ty, long long tz) {
    if (!ctxp || !pipep) return NULL;
    AmiRtMetalCtx* C = (AmiRtMetalCtx*)ctxp;
    AmiRtMetalPipe* P = (AmiRtMetalPipe*)pipep;
    id<MTLDevice> devCtx = (__bridge id<MTLDevice>)C->dev;
    id<MTLDevice> devP   = (__bridge id<MTLDevice>)P->dev;
    if (devCtx != devP) { return NULL; }
    id<MTLCommandQueue> q = (__bridge id<MTLCommandQueue>)C->q;
    id<MTLComputePipelineState> ps = (__bridge id<MTLComputePipelineState>)P->ps;
    id<MTLCommandBuffer> cb = [q commandBuffer];
    id<MTLComputeCommandEncoder> enc = [cb computeCommandEncoder];
    [enc setComputePipelineState:ps];
    // NOTE: this shim does not bind kernel args yet.
    MTLSize grid = MTLSizeMake((NSUInteger)gx, (NSUInteger)gy, (NSUInteger)gz);
    MTLSize tpg  = MTLSizeMake((NSUInteger)(tx?tx:1), (NSUInteger)(ty?ty:1), (NSUInteger)(tz?tz:1));
    [enc dispatchThreads:grid threadsPerThreadgroup:tpg];
    [enc endEncoding];
    [cb commit];
    [cb waitUntilCompleted];
    return NULL;
}

#ifdef __cplusplus
}
#endif
`), 0o644)
			metalObj = filepath.Join(rtDir, "metal_shim.o")
			// Compile Objective-C shim
			args := []string{"-x", "objective-c", "-fobjc-arc", "-fmodules", "-c", metalSrc, "-o", metalObj, "-target", triple}
			cmd := exec.Command(clang, args...)
			if outb, err := cmd.CombinedOutput(); err != nil {
				if jsonOut {
					d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_OBJ_COMPILE_FAIL", Message: "failed to compile metal shim", File: metalSrc, Data: map[string]any{"env": env, "stderr": string(outb)}}
					_ = json.NewEncoder(out).Encode(d)
				}
			}
		}
		if st, _ := os.Stat(rtObj); st != nil {
			objs = append(objs, rtObj)
		}
		// Compile generic GPU C shims (CUDA/OpenCL availability + enumeration stubs)
		{
			shim := filepath.Join(rtDir, "gpu_shims.c")
			_ = os.WriteFile(shim, []byte(`
#include <stdlib.h>
#include <stdint.h>
#include <string.h>

#if defined(_WIN32)
#include <windows.h>
static int has_cuda(void){ HMODULE h = LoadLibraryA("nvcuda.dll"); if(h){ FreeLibrary(h); return 1; } return 0; }
static int has_opencl(void){ HMODULE h = LoadLibraryA("OpenCL.dll"); if(h){ FreeLibrary(h); return 1; } return 0; }
#elif defined(__APPLE__)
#include <dlfcn.h>
static int has_cuda(void){ void* h = dlopen("libcuda.dylib", RTLD_LAZY); if(h){ dlclose(h); return 1; } return 0; }
static int has_opencl(void){ void* h = dlopen("/System/Library/Frameworks/OpenCL.framework/OpenCL", RTLD_LAZY); if(!h) h = dlopen("libOpenCL.dylib", RTLD_LAZY); if(h){ dlclose(h); return 1; } return 0; }
#else
#include <dlfcn.h>
static int has_cuda(void){ void* h = dlopen("libcuda.so.1", RTLD_LAZY); if(!h) h = dlopen("libcuda.so", RTLD_LAZY); if(h){ dlclose(h); return 1; } return 0; }
static int has_opencl(void){ void* h = dlopen("libOpenCL.so.1", RTLD_LAZY); if(!h) h = dlopen("libOpenCL.so", RTLD_LAZY); if(h){ dlclose(h); return 1; } return 0; }
#endif

// Strong definitions override weak runtime stubs
unsigned char ami_rt_cuda_available(void) {
    const char* v = getenv("AMI_GPU_FORCE_CUDA");
    if (v && v[0]) return 1;
    return has_cuda() ? 1 : 0;
}

unsigned char ami_rt_opencl_available(void) {
    const char* v = getenv("AMI_GPU_FORCE_OPENCL");
    if (v && v[0]) return 1;
    return has_opencl() ? 1 : 0;
}

// Owned handle layout: { void* data; long long len }
typedef struct { void* data; long long len; } AmiOwned;
static void* mk_owned(const char* s){
    if (!s) return (void*)0;
    size_t n = strlen(s);
    char* buf = (char*)malloc(n);
    if (!buf) return (void*)0;
    memcpy(buf, s, n);
    AmiOwned* h = (AmiOwned*)malloc(sizeof(AmiOwned));
    if (!h){ free(buf); return (void*)0; }
    h->data = (void*)buf;
    h->len = (long long)n;
    return (void*)h;
}

// Enumeration (scaffold): return JSON arrays in Owned-like handle
void* ami_rt_cuda_devices(void) {
    if (!ami_rt_cuda_available()) return (void*)0;
    // Minimal, deterministic JSON; expand when real enumeration lands
    return mk_owned("{\"backend\":\"cuda\",\"devices\":[{\"id\":0,\"name\":\"stub\"}]}\n");
}

void* ami_rt_opencl_platforms(void) {
    if (!ami_rt_opencl_available()) return (void*)0;
    return mk_owned("{\"backend\":\"opencl\",\"platforms\":[{\"name\":\"stub\"}]}\n");
}

void* ami_rt_opencl_devices(void) {
    if (!ami_rt_opencl_available()) return (void*)0;
    return mk_owned("{\"backend\":\"opencl\",\"devices\":[{\"id\":0,\"name\":\"stub\"}]}\n");
}

`), 0o644)
			shimObj := filepath.Join(rtDir, "gpu_shims.o")
			args := []string{"-x", "c", "-c", shim, "-o", shimObj, "-target", triple}
			cmd := exec.Command(clang, args...)
			if outb, err := cmd.CombinedOutput(); err != nil {
				if jsonOut {
					d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_OBJ_COMPILE_FAIL", Message: "failed to compile gpu shims", File: shim, Data: map[string]any{"env": env, "stderr": string(outb)}}
					_ = json.NewEncoder(out).Encode(d)
				}
			} else {
				if st, _ := os.Stat(shimObj); st != nil { objs = append(objs, shimObj) }
			}
		}
		if metalObj != "" {
			if st, _ := os.Stat(metalObj); st != nil {
				objs = append(objs, metalObj)
			}
		}
		// optional entry.o when no user main
		if !hasUserMain(ws, dir) {
			ingress := collectIngressIDs(ws, dir)
			if entLL, eerr := be.WriteIngressEntrypointLL(rtDir, triple, ingress); eerr == nil {
				entObj := filepath.Join(rtDir, "entry.o")
				if ecomp := be.CompileLLToObject(clang, entLL, entObj, triple); ecomp == nil {
					objs = append(objs, entObj)
				} else if jsonOut {
					d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_OBJ_COMPILE_FAIL", Message: "failed to compile LLVM to object (entry)", File: entLL, Data: map[string]any{"env": env, "what": "entry"}}
					if te, ok := ecomp.(llvme.ToolError); ok {
						if d.Data == nil {
							d.Data = map[string]any{}
						}
						d.Data["stderr"] = te.Stderr
					}
					_ = json.NewEncoder(out).Encode(d)
				}
			}
		}
		outDir := filepath.Join(dir, "build", env)
		_ = os.MkdirAll(outDir, 0o755)
		outBin := filepath.Join(outDir, binName)
		extra := linkExtraFlags(env, ws.Toolchain.Linker.Options)
		if lerr := be.LinkObjects(clang, objs, outBin, triple, extra...); lerr != nil {
			if lg := getRootLogger(); lg != nil {
				lg.Info("build.link.fail", map[string]any{"error": lerr.Error(), "bin": outBin, "env": env})
			}
			if jsonOut {
				data := map[string]any{"env": env, "bin": outBin}
				if te, ok := lerr.(llvme.ToolError); ok {
					data["stderr"] = te.Stderr
				}
				d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_LINK_FAIL", Message: "linking failed", File: "clang", Data: data}
				_ = json.NewEncoder(out).Encode(d)
			}
        } else if lg := getRootLogger(); lg != nil {
            lg.Info("build.link.ok", map[string]any{"bin": outBin, "objects": len(objs), "env": env})
        }

        // Emit workers shared library stubs for dynamic worker resolution.
        if err := emitWorkersLibs(clang, dir, ws, env, triple, jsonOut); err != nil {
            if lg := getRootLogger(); lg != nil { lg.Info("build.workerslib.fail", map[string]any{"error": err.Error(), "env": env}) }
        }
    }
}
// (emitWorkersLibs and sanitizeForCSymbol moved to separate files)
