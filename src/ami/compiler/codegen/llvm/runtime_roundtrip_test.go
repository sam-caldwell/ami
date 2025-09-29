package llvm

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// Emits a module that calls several runtime helpers (mix of captured/ignored returns),
// writes .ll, compiles with clang -c, and asserts object written.
func TestLLVM_RoundTrip_RuntimeCalls_ObjectOnly(t *testing.T) {
    clang, err := FindClang()
    if err != nil { t.Skip("clang not found; skipping") }

    // Build IR function body
    b := ir.Block{Name: "entry"}
    // sz = 4 (i64 literal)
    sz := ir.Value{ID: "sz", Type: "int64"}
    b.Instr = append(b.Instr, ir.Expr{Op: "lit:4", Result: &sz})
    // p = ami_rt_alloc(sz)
    p := ir.Value{ID: "p", Type: "ptr"}
    b.Instr = append(b.Instr, ir.Expr{Op: "call", Callee: "ami_rt_alloc", Args: []ir.Value{sz}, Result: &p})
    // call ami_rt_zeroize(p, sz) (ignored result)
    b.Instr = append(b.Instr, ir.Expr{Op: "call", Callee: "ami_rt_zeroize", Args: []ir.Value{p, sz}})
    // h = ami_rt_owned_new(p, sz)
    h := ir.Value{ID: "h", Type: "Owned"}
    b.Instr = append(b.Instr, ir.Expr{Op: "call", Callee: "ami_rt_owned_new", Args: []ir.Value{p, sz}, Result: &h})
    // call ami_rt_owned_len(h) ignored
    b.Instr = append(b.Instr, ir.Expr{Op: "call", Callee: "ami_rt_owned_len", Args: []ir.Value{h}})
    // call ami_rt_owned_ptr(h) ignored
    b.Instr = append(b.Instr, ir.Expr{Op: "call", Callee: "ami_rt_owned_ptr", Args: []ir.Value{h}})
    // call ami_rt_zeroize_owned(h)
    b.Instr = append(b.Instr, ir.Expr{Op: "call", Callee: "ami_rt_zeroize_owned", Args: []ir.Value{h}})
    // ret
    b.Instr = append(b.Instr, ir.Return{})

    f := ir.Function{Name: "F", Blocks: []ir.Block{b}}
    m := ir.Module{Package: "rt_round", Functions: []ir.Function{f}}

    ll, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }

    dir := filepath.Join("build", "test", "llvm_rt_roundtrip")
    _ = os.MkdirAll(dir, 0o755)
    llp := filepath.Join(dir, "rt.ll")
    if err := os.WriteFile(llp, []byte(ll), 0o644); err != nil { t.Fatalf("write ll: %v", err) }
    obj := filepath.Join(dir, "rt.o")
    if err := CompileLLToObject(clang, llp, obj, DefaultTriple); err != nil {
        t.Fatalf("clang compile failed: %v\nIR:\n%s", err, ll)
    }
    st, err := os.Stat(obj)
    if err != nil || st.Size() == 0 { t.Fatalf("object missing or empty: %v", err) }
}

