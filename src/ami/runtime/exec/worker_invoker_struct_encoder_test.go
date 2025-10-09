package exec

import (
    "context"
    "os/exec"
    "os"
    "path/filepath"
    "runtime"
    "testing"
    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// Build a JSON-ABI worker that uses ami_rt_value_to_json with a Struct descriptor
// and a layout of Owned pointers to confirm the encoder path is taken.
func TestWorkerInvoker_Struct_Encoder_JSONABI(t *testing.T) {
    dir := t.TempDir()
    cfile := filepath.Join(dir, "w.c")
    src := `
#include <stdlib.h>
#include <string.h>
// runtime signatures
extern void* ami_rt_owned_new(const char*, long long);
extern long long ami_rt_owned_len(void*);
extern const char* ami_rt_value_to_json(const char*, long long, int*);

// Descriptor: 'S' + i32 count(2) + field entries: len+name
static const unsigned char td[] = { 'S', 2, 0, 0, 0,  // 'S' and count=2 (little-endian)
                                    1, 'a',          // field0 name "a"
                                    1, 'b' };        // field1 name "b"

const char* ami_worker_Obj(const char* in_json, int in_len, int* out_len, const char** err){
  (void)in_json; (void)in_len; if(err) *err=NULL;
  // Values: array of two Owned pointers: a=5, b="x"
  void** vals = (void**)malloc(2*sizeof(void*));
  const char a[] = "5"; vals[0] = ami_rt_owned_new(a, (long long)1);
  const char b[] = "\"x\""; vals[1] = ami_rt_owned_new(b, (long long)3);
  long long vh = (long long)(vals);
  return ami_rt_value_to_json((const char*)td, vh, out_len);
}
`
    if err := os.WriteFile(cfile, []byte(src), 0o644); err != nil { t.Fatal(err) }
    clang, err := llvme.FindClang(); if err != nil { t.Skip("clang not found") }
    triple := llvme.DefaultTriple
    // Build runtime object
    rtDir := filepath.Join(dir, "rt")
    if _, err := llvme.WriteRuntimeLL(rtDir, triple, false); err != nil { t.Fatalf("WriteRuntimeLL: %v", err) }
    rtLL := filepath.Join(rtDir, "runtime.ll")
    rtObj := filepath.Join(rtDir, "runtime.o")
    if err := llvme.CompileLLToObject(clang, rtLL, rtObj, triple); err != nil { t.Skipf("compile runtime.ll failed: %v", err) }
    // Link shared lib
    var lib string
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "linux": lib = filepath.Join(dir, "libw.so"); cmd = exec.Command(clang, "-shared", "-fPIC", cfile, rtObj, "-o", lib, "-target", triple)
    case "darwin": lib = filepath.Join(dir, "libw.dylib"); cmd = exec.Command(clang, "-dynamiclib", cfile, rtObj, "-o", lib, "-target", triple)
    default: t.Skip("unsupported OS")
    }
    if out, err := cmd.CombinedOutput(); err != nil { t.Skipf("compile/link failed: %v, out=%s", err, string(out)) }
    inv := NewDLSOInvoker(lib, "ami_worker_")
    if inv == nil { t.Skip("invoker unavailable") }
    call, ok := inv.Resolve("Obj"); if !ok || call == nil { t.Fatalf("resolve Obj failed") }
    ctx, cancel := context.WithCancel(context.Background()); defer cancel(); _ = ctx
    v, err := call(ev.Event{Payload: 1}); if err != nil { t.Fatalf("call: %v", err) }
    m, ok := v.(map[string]any); if !ok { t.Fatalf("expected object, got %T", v) }
    if int(m["a"].(float64)) != 5 { t.Fatalf("a=%v", m["a"]) }
    if s, _ := m["b"].(string); s != "x" { t.Fatalf("b=%v", m["b"]) }
}

