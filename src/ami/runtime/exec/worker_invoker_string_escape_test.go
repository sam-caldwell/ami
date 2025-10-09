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

// Build a JSON-ABI worker that uses runtime string->JSON to escape tricky strings, then validate result via DLSOInvoker.
func TestWorkerInvoker_StringEscaping_JSONABI(t *testing.T) {
    dir := t.TempDir()
    cfile := filepath.Join(dir, "w.c")
    // The input bytes include quotes, backslash, newline, tab, carriage, backspace, formfeed, and 0x01
    src := `
#include <stdlib.h>
#include <string.h>

// runtime signatures
extern void* ami_rt_owned_new(const char*, long long);
extern const char* ami_rt_string_to_json(void*, int*);

static const unsigned char tricky[] = {
  'a','"','b','\\','c','\n','\t','\r','\b','\f',
  0x00, 0x01, 0x02, 0x07, 0x0B, 0x0E, 0x1F
};

const char* ami_worker_Str(const char* in_json, int in_len, int* out_len, const char** err){
  (void)in_json; (void)in_len; if(err) *err=NULL;
  void* h = ami_rt_owned_new((const char*)tricky, (long long)sizeof(tricky));
  return ami_rt_string_to_json(h, out_len);
}
`
    if err := os.WriteFile(cfile, []byte(src), 0o644); err != nil { t.Fatal(err) }
    clang, err := llvme.FindClang(); if err != nil { t.Skip("clang not found") }
    triple := llvme.DefaultTriple
    // Build runtime.ll object
    rtDir := filepath.Join(dir, "rt")
    if _, err := llvme.WriteRuntimeLL(rtDir, triple, false); err != nil { t.Fatalf("WriteRuntimeLL: %v", err) }
    rtLL := filepath.Join(rtDir, "runtime.ll")
    rtObj := filepath.Join(rtDir, "runtime.o")
    if err := llvme.CompileLLToObject(clang, rtLL, rtObj, triple); err != nil { t.Skipf("compile runtime.ll failed: %v", err) }
    var lib string
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "linux":
        lib = filepath.Join(dir, "libw.so")
        cmd = exec.Command(clang, "-shared", "-fPIC", cfile, rtObj, "-o", lib, "-target", triple)
    case "darwin":
        lib = filepath.Join(dir, "libw.dylib")
        cmd = exec.Command(clang, "-dynamiclib", cfile, rtObj, "-o", lib, "-target", triple)
    default:
        t.Skip("dynamic linking test not supported on this OS")
    }
    if out, err := cmd.CombinedOutput(); err != nil { t.Skipf("compile failed: %v, out=%s", err, string(out)) }

    inv := NewDLSOInvoker(lib, "ami_worker_")
    if inv == nil { t.Skip("invoker unavailable (no cgo)") }
    call, ok := inv.Resolve("Str")
    if !ok || call == nil { t.Fatalf("resolve Str failed") }
    ctx, cancel := context.WithCancel(context.Background()); defer cancel(); _ = ctx
    v, err := call(ev.Event{Payload: 1})
    if err != nil { t.Fatalf("call: %v", err) }
    s, ok := v.(string)
    if !ok { t.Fatalf("expected string, got %T", v) }
    want := []byte{'a','"','b','\\','c','\n','\t','\r','\b','\f', 0x00, 0x01, 0x02, 0x07, 0x0B, 0x0E, 0x1F}
    if len(s) != len(want) { t.Fatalf("length mismatch: got=%d want=%d", len(s), len(want)) }
    for i := range want { if s[i] != want[i] { t.Fatalf("byte %d mismatch: got=%d want=%d", i, s[i], want[i]) } }
}
