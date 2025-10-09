package exec

import (
    "encoding/json"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "testing"

    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

// Test compiles runtime.ll and a toy workers.ll (with a worker-shaped function W)
// into a shared library that exports ami_worker_W via a tiny C wrapper calling
// ami_worker_core_W. Then it invokes the worker via DLSOInvoker and validates that
// the JSON body returned is composed as prefix + input + suffix, where the payload
// mirrors the input JSON bytes per the current minimal JSON bridge.
func TestWorkerCoreLLVM_RoundTrip_SharedLib(t *testing.T) {
    clang, err := llvme.FindClang()
    if err != nil { t.Skip("clang not found; skipping") }
    if ver, err := llvme.Version(clang); err == nil && ver == "" {
        t.Skip("clang version undetectable; skipping")
    }
    triple := llvme.DefaultTriple
    dir := t.TempDir()

    // 1) Write runtime.ll
    rtDir := filepath.Join(dir, "rt")
    if _, err := llvme.WriteRuntimeLL(rtDir, triple, false); err != nil { t.Fatalf("WriteRuntimeLL: %v", err) }
    rtLL := filepath.Join(rtDir, "runtime.ll")
    rtObj := filepath.Join(rtDir, "runtime.o")
    if err := llvme.CompileLLToObject(clang, rtLL, rtObj, triple); err != nil { t.Skipf("compile runtime.ll failed: %v", err) }

    // 2) Build workers.ll from an IR module containing W(ev Event)->(Event, error)
    // W returns its input event and a nil error; implement a small Err() helper
    // that defaults to returning a null pointer via the emitter's default path.
    wFn := ir.Function{
        Name:   "W",
        Params: []ir.Value{{ID: "ev", Type: "Event<int>"}},
        Results: []ir.Value{{ID: "r0", Type: "Event<int>"}, {ID: "r1", Type: "error"}},
        Blocks: []ir.Block{{
            Name: "entry",
            Instr: []ir.Instruction{
                ir.Expr{Op: "call", Callee: "Err", Result: &ir.Value{ID: "e", Type: "error"}},
                ir.Return{Values: []ir.Value{{ID: "ev", Type: "Event<int>"}, {ID: "e", Type: "error"}}},
            },
        }},
    }
    errFn := ir.Function{Name: "Err", Results: []ir.Value{{ID: "e", Type: "error"}}}
    mod := ir.Module{Package: "app", Functions: []ir.Function{wFn, errFn}}
    wLL, err := llvme.EmitModuleLLVMForTarget(mod, triple)
    if err != nil { t.Fatalf("EmitModuleLLVMForTarget: %v", err) }
    wDir := filepath.Join(dir, "w")
    if err := os.MkdirAll(wDir, 0o755); err != nil { t.Fatal(err) }
    wLLPath := filepath.Join(wDir, "workers.ll")
    if err := os.WriteFile(wLLPath, []byte(wLL), 0o644); err != nil { t.Fatal(err) }
    wObj := filepath.Join(wDir, "workers.o")
    if err := llvme.CompileLLToObject(clang, wLLPath, wObj, triple); err != nil { t.Skipf("compile workers.ll failed: %v", err) }

    // 3) Provide a tiny C wrapper that exports ami_worker_W and calls ami_worker_core_W
    cSrc := `
#include <stdlib.h>
typedef const char* (*core_fn_t)(const char*, int, int*, const char**);
extern const char* ami_worker_core_W(const char*, int, int*, const char**);
const char* ami_worker_W(const char* in_json, int in_len, int* out_len, const char** err){
    core_fn_t f = (core_fn_t)ami_worker_core_W; return f(in_json, in_len, out_len, err);
}
`
    cPath := filepath.Join(dir, "wrap.c")
    if err := os.WriteFile(cPath, []byte(cSrc), 0o644); err != nil { t.Fatal(err) }
    cObj := filepath.Join(dir, "wrap.o")
    // Compile C wrapper for the same target
    if out, err := exec.Command(clang, "-c", cPath, "-o", cObj, "-target", triple).CombinedOutput(); err != nil {
        t.Skipf("compile wrapper.c failed: %v, out=%s", err, string(out))
    }

    // 4) Link into a shared library
    var lib string
    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "linux":
        lib = filepath.Join(dir, "libw.so")
        cmd = exec.Command(clang, "-shared", "-fPIC", cObj, rtObj, wObj, "-o", lib, "-target", triple)
    case "darwin":
        lib = filepath.Join(dir, "libw.dylib")
        cmd = exec.Command(clang, "-dynamiclib", cObj, rtObj, wObj, "-o", lib, "-target", triple)
    default:
        t.Skip("OS not supported for shared linking")
    }
    if out, err := cmd.CombinedOutput(); err != nil {
        t.Skipf("link shared lib failed: %v, out=%s", err, string(out))
    }

    // 5) Resolve and invoke via DLSOInvoker
    inv := NewDLSOInvoker(lib, "ami_worker_")
    if inv == nil { t.Skip("invoker unavailable (no cgo)") }
    call, ok := inv.Resolve("W")
    if !ok || call == nil { t.Fatalf("failed to resolve worker symbol W") }
    // Input event with simple payload
    inEv := ev.Event{Payload: map[string]any{"x": 7}}
    // Marshal input to build expected output JSON
    inJSON, _ := json.Marshal(inEv)
    outAny, err := call(inEv)
    if err != nil { t.Fatalf("invocation error: %v", err) }
    outEv, ok := outAny.(ev.Event)
    if !ok { t.Fatalf("expected Event output, got %T", outAny) }
    gotJSON, _ := json.Marshal(outEv)
    // Expect prefix + input + suffix composition
    exp := append([]byte("{\"schema\":\"events.v1\",\"payload\":"), inJSON...)
    exp = append(exp, byte('}'))
    if string(gotJSON) != string(exp) {
        t.Fatalf("unexpected output JSON:\n got=%s\nexp=%s", string(gotJSON), string(exp))
    }
}

