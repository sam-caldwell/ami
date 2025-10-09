package exec

import (
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "testing"
    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// Build wrappers-only LLVM for four array workers (slice<int>, slice<float64>, slice<bool>, slice<string>),
// provide C bodies that construct header-mode arrays with raw pointers for i/d/b and Owned strings for s,
// link with runtime, invoke via DLSOInvoker, and verify JSON decoding of payload arrays.
func TestCoreWrapper_Array_Kinds_HeaderMode(t *testing.T) {
    clang, err := llvme.FindClang(); if err != nil { t.Skip("clang not found") }
    triple := llvme.DefaultTriple
    dir := t.TempDir()

    // IR module with 4 worker-shaped signatures
    fni := ir.Function{Name: "Ai", Params: []ir.Value{{ID: "ev", Type: "Event<slice<int>>"}}, Results: []ir.Value{{ID: "r0", Type: "slice<int>"}, {ID: "r1", Type: "error"}}}
    fnd := ir.Function{Name: "Ad", Params: []ir.Value{{ID: "ev", Type: "Event<slice<float64>>"}}, Results: []ir.Value{{ID: "r0", Type: "slice<float64>"}, {ID: "r1", Type: "error"}}}
    fnb := ir.Function{Name: "Ab", Params: []ir.Value{{ID: "ev", Type: "Event<slice<bool>>"}}, Results: []ir.Value{{ID: "r0", Type: "slice<bool>"}, {ID: "r1", Type: "error"}}}
    fns := ir.Function{Name: "As", Params: []ir.Value{{ID: "ev", Type: "Event<slice<string>>"}}, Results: []ir.Value{{ID: "r0", Type: "slice<string>"}, {ID: "r1", Type: "error"}}}
    m := ir.Module{Package: "app", Functions: []ir.Function{fni, fnd, fnb, fns}}
    // Emit wrappers-only and compile
    wLL, err := llvme.EmitWorkerWrappersOnlyForTarget(m, triple)
    if err != nil { t.Fatalf("emit wrappers: %v", err) }
    wDir := filepath.Join(dir, "wrap")
    _ = os.MkdirAll(wDir, 0o755)
    wLLPath := filepath.Join(wDir, "wrap.ll")
    if err := os.WriteFile(wLLPath, []byte(wLL), 0o644); err != nil { t.Fatal(err) }
    wObj := filepath.Join(wDir, "wrap.o")
    if err := llvme.CompileLLToObject(clang, wLLPath, wObj, triple); err != nil { t.Skipf("compile wrapper ll failed: %v", err) }

    // Runtime object
    rtDir := filepath.Join(dir, "rt")
    if _, err := llvme.WriteRuntimeLL(rtDir, triple, false); err != nil { t.Fatalf("WriteRuntimeLL: %v", err) }
    rtLL := filepath.Join(rtDir, "runtime.ll")
    rtObj := filepath.Join(rtDir, "runtime.o")
    if err := llvme.CompileLLToObject(clang, rtLL, rtObj, triple); err != nil { t.Skipf("compile runtime.ll failed: %v", err) }

    // C bodies: each W builds header {count, elems} and returns (handle,0). Also shim ami_worker_X calls core wrapper.
    cSrc := `
#include <stdlib.h>
typedef struct { long long r0; long long r1; } pair64;
typedef const char* (*core_t)(const char*, int, int*, const char**);
extern const char* ami_worker_core_Ai(const char*, int, int*, const char**);
extern const char* ami_worker_core_Ad(const char*, int, int*, const char**);
extern const char* ami_worker_core_Ab(const char*, int, int*, const char**);
extern const char* ami_worker_core_As(const char*, int, int*, const char**);
extern void* ami_rt_owned_new(const char*, long long);

// Helper: header with count and pointer to elems array
static void* hdr(void** elems, long long n){ long long* h=(long long*)malloc(16); h[0]=n; *(void**)(&h[1])=(void*)elems; return (void*)h; }

pair64 Ai(long long ev){ (void)ev; long long *a=(long long*)malloc(2*sizeof(long long)); a[0]=5; a[1]=7; void** elems=(void**)malloc(2*sizeof(void*)); elems[0]=&a[0]; elems[1]=&a[1]; void* h=hdr(elems,2); pair64 o; o.r0=(long long)h; o.r1=0; return o; }
pair64 Ad(long long ev){ (void)ev; double *a=(double*)malloc(2*sizeof(double)); a[0]=3.5; a[1]=4.25; void** elems=(void**)malloc(2*sizeof(void*)); elems[0]=&a[0]; elems[1]=&a[1]; void* h=hdr(elems,2); pair64 o; o.r0=(long long)h; o.r1=0; return o; }
pair64 Ab(long long ev){ (void)ev; unsigned char *a=(unsigned char*)malloc(3); a[0]=1; a[1]=0; a[2]=1; void** elems=(void**)malloc(3*sizeof(void*)); elems[0]=&a[0]; elems[1]=&a[1]; elems[2]=&a[2]; void* h=hdr(elems,3); pair64 o; o.r0=(long long)h; o.r1=0; return o; }
pair64 As(long long ev){ (void)ev; void** elems=(void**)malloc(2*sizeof(void*)); const char s0[]="\"x\""; const char s1[]="\"y\""; elems[0]=ami_rt_owned_new(s0,3); elems[1]=ami_rt_owned_new(s1,3); void* h=hdr(elems,2); pair64 o; o.r0=(long long)h; o.r1=0; return o; }

const char* ami_worker_Ai(const char* in, int inlen, int* outlen, const char** err){ core_t f=(core_t)ami_worker_core_Ai; return f(in,inlen,outlen,err); }
const char* ami_worker_Ad(const char* in, int inlen, int* outlen, const char** err){ core_t f=(core_t)ami_worker_core_Ad; return f(in,inlen,outlen,err); }
const char* ami_worker_Ab(const char* in, int inlen, int* outlen, const char** err){ core_t f=(core_t)ami_worker_core_Ab; return f(in,inlen,outlen,err); }
const char* ami_worker_As(const char* in, int inlen, int* outlen, const char** err){ core_t f=(core_t)ami_worker_core_As; return f(in,inlen,outlen,err); }
`
    cPath := filepath.Join(dir, "arr.c")
    if err := os.WriteFile(cPath, []byte(cSrc), 0o644); err != nil { t.Fatal(err) }

    // Link shared library
    var lib string
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "linux": lib = filepath.Join(dir, "libarr.so"); cmd = exec.Command(clang, "-shared", "-fPIC", cPath, wObj, rtObj, "-o", lib, "-target", triple)
    case "darwin": lib = filepath.Join(dir, "libarr.dylib"); cmd = exec.Command(clang, "-dynamiclib", cPath, wObj, rtObj, "-o", lib, "-target", triple)
    default: t.Skip("unsupported OS")
    }
    if out, err := cmd.CombinedOutput(); err != nil { t.Skipf("link failed: %v, out=%s", err, string(out)) }

    inv := NewDLSOInvoker(lib, "ami_worker_")
    if inv == nil { t.Skip("invoker unavailable") }
    ctx, cancel := context.WithCancel(context.Background()); defer cancel(); _ = ctx

    // Helper to call and decode payload as []any
    callArr := func(name string) []any {
        f, ok := inv.Resolve(name); if !ok || f == nil { t.Fatalf("resolve %s failed", name) }
        v, err := f(ev.Event{Payload: 1}); if err != nil { t.Fatalf("%s: %v", name, err) }
        arr, ok := v.([]any); if !ok { t.Fatalf("%s not array, got %T", name, v) }
        return arr
    }
    ai := callArr("Ai"); if len(ai)!=2 || int(ai[0].(float64))!=5 || int(ai[1].(float64))!=7 { t.Fatalf("Ai=%v", ai) }
    ad := callArr("Ad"); if len(ad)!=2 || ad[0].(float64)!=3.5 || ad[1].(float64)!=4.25 { t.Fatalf("Ad=%v", ad) }
    ab := callArr("Ab"); if len(ab)!=3 || ab[0].(bool)!=true || ab[1].(bool)!=false || ab[2].(bool)!=true { t.Fatalf("Ab=%v", ab) }
    as := callArr("As"); if len(as)!=2 || as[0].(string)!="x" || as[1].(string)!="y" { t.Fatalf("As=%v", as) }
}

