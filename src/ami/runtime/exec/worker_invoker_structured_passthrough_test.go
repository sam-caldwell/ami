package exec

import (
    "context"
    "os/exec"
    "os"
    "path/filepath"
    "runtime"
    "testing"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// Build a JSON-ABI worker that uses runtime structured->JSON passthrough from an Owned buffer.
func TestWorkerInvoker_Structured_Passthrough_JSONABI(t *testing.T) {
    dir := t.TempDir()
    cfile := filepath.Join(dir, "w.c")
    src := `
#include <stdlib.h>
#include <string.h>
extern void* ami_rt_owned_new(const char*, long long);
extern const char* ami_rt_structured_to_json(void*, int*);
static const char json[] = "{\"a\":5,\"b\":\"x\"}";
const char* ami_worker_Struct(const char* in_json, int in_len, int* out_len, const char** err){
  (void)in_json; (void)in_len; if(err) *err=NULL;
  void* h = ami_rt_owned_new(json, (long long)(sizeof(json)-1));
  return ami_rt_structured_to_json(h, out_len);
}
`
    if err := os.WriteFile(cfile, []byte(src), 0o644); err != nil { t.Fatal(err) }
    var lib string
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "linux": lib = filepath.Join(dir, "libw.so"); cmd = exec.Command("cc", "-shared", "-fPIC", cfile, "-o", lib)
    case "darwin": lib = filepath.Join(dir, "libw.dylib"); cmd = exec.Command("cc", "-dynamiclib", cfile, "-o", lib)
    default: t.Skip("dynamic linking test not supported on this OS")
    }
    if out, err := cmd.CombinedOutput(); err != nil { t.Skipf("compile failed: %v, out=%s", err, string(out)) }

    inv := NewDLSOInvoker(lib, "ami_worker_")
    if inv == nil { t.Skip("invoker unavailable (no cgo)") }
    call, ok := inv.Resolve("Struct")
    if !ok || call == nil { t.Fatalf("resolve Struct failed") }
    ctx, cancel := context.WithCancel(context.Background()); defer cancel(); _ = ctx
    v, err := call(ev.Event{Payload: map[string]any{"i":1}})
    if err != nil { t.Fatalf("call: %v", err) }
    m, ok := v.(map[string]any); if !ok { t.Fatalf("expected map payload, got %T", v) }
    if int(m["a"].(float64)) != 5 { t.Fatalf("unexpected a: %v", m["a"]) }
    if s, _ := m["b"].(string); s != "x" { t.Fatalf("unexpected b: %q", s) }
}

