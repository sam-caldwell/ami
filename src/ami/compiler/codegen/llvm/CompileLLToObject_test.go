package llvm

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestCompileLLToObject_WithClangOrSkip(t *testing.T) {
    clang, err := FindClang()
    if err != nil {
        t.Skip("clang not found; skipping object compilation test")
    }
    // Build a tiny module and write to temp dir under build/test
    m := ir.Module{Package: "objtest", Functions: []ir.Function{{Name: "Foo", Results: []ir.Value{{Type: "int"}}, Blocks: []ir.Block{{Name: "entry", Instr: nil}}}}}
    ll, err := EmitModuleLLVM(m)
    if err != nil { t.Fatalf("emit: %v", err) }
    dir := filepath.Join("build", "test", "llvm_obj")
    _ = os.MkdirAll(dir, 0o755)
    llp := filepath.Join(dir, "m.ll")
    if err := os.WriteFile(llp, []byte(ll), 0o644); err != nil { t.Fatalf("write ll: %v", err) }
    out := filepath.Join(dir, "m.o")
    if err := CompileLLToObject(clang, llp, out, DefaultTriple); err != nil {
        t.Fatalf("clang compile failed: %v", err)
    }
    st, err := os.Stat(out)
    if err != nil || st.Size() == 0 { t.Fatalf("object not written or empty: %v", err) }
}
