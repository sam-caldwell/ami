package exec

import (
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "testing"
    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// slice of structs: array walker should use nested struct descriptor embedded in the slice descriptor.
func TestCoreWrapper_ArrayOfStructs_NestedEncoder(t *testing.T) {
    clang, err := llvme.FindClang(); if err != nil { t.Skip("clang not found") }
    triple := llvme.DefaultTriple
    dir := t.TempDir()
    // IR: A(Event<slice<Struct{a:int,b:string}>>) -> (slice<Struct{a:int,b:string}>, error)
    fn := ir.Function{Name: "A", Params: []ir.Value{{ID: "ev", Type: "Event<slice<Struct{a:int,b:string}>>"}}, Results: []ir.Value{{ID: "r0", Type: "slice<Struct{a:int,b:string}>"}, {ID: "r1", Type: "error"}}}
    m := ir.Module{Package: "app", Functions: []ir.Function{fn}}
    wLL, err := llvme.EmitWorkerWrappersOnlyForTarget(m, triple); if err != nil { t.Fatalf("emit: %v", err) }
    wDir := filepath.Join(dir, "wrap"); _ = os.MkdirAll(wDir, 0o755)
    wLLPath := filepath.Join(wDir, "wrap.ll"); _ = os.WriteFile(wLLPath, []byte(wLL), 0o644)
    wObj := filepath.Join(wDir, "wrap.o"); if err := llvme.CompileLLToObject(clang, wLLPath, wObj, triple); err != nil { t.Skipf("compile wrap: %v", err) }
    // runtime
    rtDir := filepath.Join(dir, "rt"); if _, err := llvme.WriteRuntimeLL(rtDir, triple, false); err != nil { t.Fatalf("rt: %v", err) }
    rtObj := filepath.Join(rtDir, "runtime.o"); if err := llvme.CompileLLToObject(clang, filepath.Join(rtDir, "runtime.ll"), rtObj, triple); err != nil { t.Skipf("compile rt: %v", err) }
    // C W constructs header-mode slice of 2 structs: [{a:5,b:"x"},{a:7,b:"y"}]
    cSrc := `
#include <stdlib.h>
typedef struct { long long r0; long long r1; } pair64;
extern void* ami_rt_owned_new(const char*, long long);
pair64 A(long long ev){ (void)ev; void** elems=(void**)malloc(2*sizeof(void*));
  long long *a0=(long long*)malloc(sizeof(long long)); *a0=5; void** sf0=(void**)malloc(2*sizeof(void*)); sf0[0]=a0; sf0[1]=ami_rt_owned_new("\"x\"",3);
  long long *a1=(long long*)malloc(sizeof(long long)); *a1=7; void** sf1=(void**)malloc(2*sizeof(void*)); sf1[0]=a1; sf1[1]=ami_rt_owned_new("\"y\"",3);
  elems[0]=sf0; elems[1]=sf1; long long* hdr=(long long*)malloc(16); hdr[0]=2; *(void**)(&hdr[1])=(void*)elems; pair64 o; o.r0=(long long)hdr; o.r1=0; return o; }
typedef const char* (*core_t)(const char*, int, int*, const char**);
extern const char* ami_worker_core_A(const char*, int, int*, const char**);
const char* ami_worker_A(const char* in, int inlen, int* outlen, const char** err){ core_t f=(core_t)ami_worker_core_A; return f(in,inlen,outlen,err); }
`
    cPath := filepath.Join(dir, "w.c"); _ = os.WriteFile(cPath, []byte(cSrc), 0o644)
    var lib string; var cmd *exec.Cmd
    switch runtime.GOOS {
    case "linux": lib=filepath.Join(dir, "libarrs.so"); cmd=exec.Command(clang, "-shared", "-fPIC", cPath, wObj, rtObj, "-o", lib, "-target", triple)
    case "darwin": lib=filepath.Join(dir, "libarrs.dylib"); cmd=exec.Command(clang, "-dynamiclib", cPath, wObj, rtObj, "-o", lib, "-target", triple)
    default: t.Skip("unsupported OS")
    }
    if out, err := cmd.CombinedOutput(); err != nil { t.Skipf("link failed: %v, out=%s", err, string(out)) }
    inv := NewDLSOInvoker(lib, "ami_worker_"); if inv == nil { t.Skip("no invoker") }
    call, ok := inv.Resolve("A"); if !ok || call == nil { t.Fatalf("resolve A failed") }
    v, err := call(ev.Event{Payload: 1}); if err != nil { t.Fatalf("call: %v", err) }
    arr, ok := v.([]any); if !ok || len(arr)!=2 { t.Fatalf("want 2 structs, got %T %v", v, v) }
    m0, ok := arr[0].(map[string]any); if !ok { t.Fatalf("elem0 type: %T", arr[0]) }
    m1, ok := arr[1].(map[string]any); if !ok { t.Fatalf("elem1 type: %T", arr[1]) }
    if int(m0["a"].(float64))!=5 || m0["b"].(string)!="x" { t.Fatalf("elem0=%v", m0) }
    if int(m1["a"].(float64))!=7 || m1["b"].(string)!="y" { t.Fatalf("elem1=%v", m1) }
}
