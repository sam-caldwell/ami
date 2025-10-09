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

// Build a shared lib containing three payload-return workers and verify JSON bridging for primitives.
func TestWorkerCoreLLVM_RoundTrip_PayloadPrimitives(t *testing.T) {
    clang, err := llvme.FindClang()
    if err != nil { t.Skip("clang not found; skipping") }
    if ver, err := llvme.Version(clang); err == nil && ver == "" { t.Skip("clang version undetectable; skipping") }
    triple := llvme.DefaultTriple
    dir := t.TempDir()

    // Runtime
    rtDir := filepath.Join(dir, "rt")
    if _, err := llvme.WriteRuntimeLL(rtDir, triple, false); err != nil { t.Fatalf("WriteRuntimeLL: %v", err) }
    rtLL := filepath.Join(rtDir, "runtime.ll")
    rtObj := filepath.Join(rtDir, "runtime.o")
    if err := llvme.CompileLLToObject(clang, rtLL, rtObj, triple); err != nil { t.Skipf("compile runtime.ll failed: %v", err) }

    // Workers IR: Wi -> (int,error)=42,nil; Wb -> (bool,error)=true,nil; Wf -> (real,error)=7.0,nil
    mkWorker := func(name, retType, lit string) ir.Function {
        // value result id is same as name to keep unique
        valID := "v_" + name
        var blk ir.Block
        // literal instruction
        blk.Instr = append(blk.Instr, ir.Expr{Op: "lit:" + lit, Result: &ir.Value{ID: valID, Type: retType}})
        // error via Err (returns null by default)
        blk.Instr = append(blk.Instr, ir.Expr{Op: "call", Callee: "Err", Result: &ir.Value{ID: "e_" + name, Type: "error"}})
        // return both
        blk.Instr = append(blk.Instr, ir.Return{Values: []ir.Value{{ID: valID, Type: retType}, {ID: "e_" + name, Type: "error"}}})
        return ir.Function{Name: name, Params: []ir.Value{{ID: "ev", Type: "Event<int>"}}, Results: []ir.Value{{ID: "r0", Type: retType}, {ID: "r1", Type: "error"}}, Blocks: []ir.Block{blk}}
    }
    fWi := mkWorker("Wi", "int", "42")
    fWb := mkWorker("Wb", "bool", "1")
    fWf := mkWorker("Wf", "real", "7")
    fErr := ir.Function{Name: "Err", Results: []ir.Value{{ID: "e", Type: "error"}}}
    mod := ir.Module{Package: "app", Functions: []ir.Function{fWi, fWb, fWf, fErr}}
    wLL, err := llvme.EmitModuleLLVMForTarget(mod, triple)
    if err != nil { t.Fatalf("EmitModuleLLVMForTarget: %v", err) }
    wDir := filepath.Join(dir, "w")
    if err := os.MkdirAll(wDir, 0o755); err != nil { t.Fatal(err) }
    wLLPath := filepath.Join(wDir, "workers.ll")
    if err := os.WriteFile(wLLPath, []byte(wLL), 0o644); err != nil { t.Fatal(err) }
    wObj := filepath.Join(wDir, "workers.o")
    if err := llvme.CompileLLToObject(clang, wLLPath, wObj, triple); err != nil { t.Skipf("compile workers.ll failed: %v", err) }

    // C wrappers
    cSrc := `
#include <stdlib.h>
typedef const char* (*core_fn_t)(const char*, int, int*, const char**);
extern const char* ami_worker_core_Wi(const char*, int, int*, const char**);
extern const char* ami_worker_core_Wb(const char*, int, int*, const char**);
extern const char* ami_worker_core_Wf(const char*, int, int*, const char**);
const char* ami_worker_Wi(const char* in_json, int in_len, int* out_len, const char** err){ core_fn_t f = (core_fn_t)ami_worker_core_Wi; return f(in_json, in_len, out_len, err); }
const char* ami_worker_Wb(const char* in_json, int in_len, int* out_len, const char** err){ core_fn_t f = (core_fn_t)ami_worker_core_Wb; return f(in_json, in_len, out_len, err); }
const char* ami_worker_Wf(const char* in_json, int in_len, int* out_len, const char** err){ core_fn_t f = (core_fn_t)ami_worker_core_Wf; return f(in_json, in_len, out_len, err); }
`
    cPath := filepath.Join(dir, "wrap.c")
    if err := os.WriteFile(cPath, []byte(cSrc), 0o644); err != nil { t.Fatal(err) }
    cObj := filepath.Join(dir, "wrap.o")
    if out, err := exec.Command(clang, "-c", cPath, "-o", cObj, "-target", triple).CombinedOutput(); err != nil {
        t.Skipf("compile wrapper.c failed: %v, out=%s", err, string(out))
    }

    // Link shared library
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
    if out, err := cmd.CombinedOutput(); err != nil { t.Skipf("link shared lib failed: %v, out=%s", err, string(out)) }

    inv := NewDLSOInvoker(lib, "ami_worker_")
    if inv == nil { t.Skip("invoker unavailable (no cgo)") }
    // Common input event
    inEv := ev.Event{Payload: 1}

    // Wi -> 42
    if f, ok := inv.Resolve("Wi"); !ok || f == nil { t.Fatalf("resolve Wi failed") } else {
        v, err := f(inEv); if err != nil { t.Fatalf("Wi: %v", err) }
        switch x := v.(type) {
        case float64: if int(x) != 42 { t.Fatalf("Wi got %v", v) }
        case int: if x != 42 { t.Fatalf("Wi got %v", v) }
        default: t.Fatalf("Wi unexpected type %T", v)
        }
    }
    // Wb -> true
    if f, ok := inv.Resolve("Wb"); !ok || f == nil { t.Fatalf("resolve Wb failed") } else {
        v, err := f(inEv); if err != nil { t.Fatalf("Wb: %v", err) }
        b, ok := v.(bool); if !ok || !b { t.Fatalf("Wb got %v", v) }
    }
    // Wf -> 7.0
    if f, ok := inv.Resolve("Wf"); !ok || f == nil { t.Fatalf("resolve Wf failed") } else {
        v, err := f(inEv); if err != nil { t.Fatalf("Wf: %v", err) }
        switch x := v.(type) {
        case float64: if x != 7.0 { t.Fatalf("Wf got %v", v) }
        default: t.Fatalf("Wf unexpected type %T", v)
        }
    }
}

