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

// Build an LLVM worker-core wrapper module (wrappers only), link it with a C definition of W and runtime.ll,
// export ami_worker_W -> ami_worker_core_W, then invoke via DLSOInvoker to exercise structured encoder path.
func TestCoreWrapper_Structured_EndToEnd(t *testing.T) {
    clang, err := llvme.FindClang(); if err != nil { t.Skip("clang not found") }
    triple := llvme.DefaultTriple
    dir := t.TempDir()

    // IR module with a worker-shaped signature: W(Event<Struct{a:int,b:string}>) -> (Struct{a:int,b:string}, error)
    fn := ir.Function{Name: "W", Params: []ir.Value{{ID: "ev", Type: "Event<Struct{a:int,b:string}>"}}, Results: []ir.Value{{ID: "r0", Type: "Struct{a:int,b:string}"}, {ID: "r1", Type: "error"}}}
    m := ir.Module{Package: "app", Functions: []ir.Function{fn}}
    // Emit wrappers-only LLVM
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

    // C definition of W and ami_worker_W shim
    cSrc := `
#include <stdlib.h>
typedef struct { long long r0; long long r1; } pair64;
extern void* ami_rt_owned_new(const char*, long long);
// W returns (handle, error) where handle points to array of two Owned pointers for fields a and b.
pair64 W(long long ev){ (void)ev; void** vals = (void**)malloc(2*sizeof(void*)); const char a[]="5"; vals[0]=ami_rt_owned_new(a,1); const char b[]="\"x\""; vals[1]=ami_rt_owned_new(b,3); pair64 out; out.r0=(long long)vals; out.r1=0; return out; }
typedef const char* (*core_t)(const char*, int, int*, const char**);
extern const char* ami_worker_core_W(const char*, int, int*, const char**);
const char* ami_worker_W(const char* in, int inlen, int* outlen, const char** err){ core_t f = (core_t)ami_worker_core_W; return f(in, inlen, outlen, err); }
`
    cPath := filepath.Join(dir, "w.c")
    if err := os.WriteFile(cPath, []byte(cSrc), 0o644); err != nil { t.Fatal(err) }

    // Link shared lib: runtime.o + wrapper.o + c W file
    var lib string
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "linux":
        lib = filepath.Join(dir, "libw.so")
        cmd = exec.Command(clang, "-shared", "-fPIC", cPath, wObj, rtObj, "-o", lib, "-target", triple)
    case "darwin":
        lib = filepath.Join(dir, "libw.dylib")
        cmd = exec.Command(clang, "-dynamiclib", cPath, wObj, rtObj, "-o", lib, "-target", triple)
    default:
        t.Skip("unsupported OS for shared lib")
    }
    if out, err := cmd.CombinedOutput(); err != nil { t.Skipf("link failed: %v, out=%s", err, string(out)) }

    inv := NewDLSOInvoker(lib, "ami_worker_")
    if inv == nil { t.Skip("no invoker") }
    call, ok := inv.Resolve("W"); if !ok || call == nil { t.Fatalf("resolve W failed") }
    ctx, cancel := context.WithCancel(context.Background()); defer cancel(); _ = ctx
    v, err := call(ev.Event{Payload: map[string]any{"i":1}}); if err != nil { t.Fatalf("call: %v", err) }
    mOut, ok := v.(map[string]any); if !ok { t.Fatalf("expected map payload, got %T", v) }
    if int(mOut["a"].(float64)) != 5 { t.Fatalf("a=%v", mOut["a"]) }
    if s, _ := mOut["b"].(string); s != "x" { t.Fatalf("b=%v", mOut["b"]) }
}

