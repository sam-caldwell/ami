package codegen

import (
    "os"
    "path/filepath"
    "testing"
    ir "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func TestLLVMBackend_WrapperMethods_SafePaths(t *testing.T) {
    b := &llvmBackend{}
    if b.Name() != "llvm" { t.Fatalf("name: %s", b.Name()) }
    // Emit simple empty module
    ll, err := b.EmitModule(ir.Module{Package: "app"})
    if err != nil || ll == "" { t.Fatalf("EmitModule: %v %q", err, ll) }
    ll2, err := b.EmitModuleForTarget(ir.Module{Package: "app"}, "x86_64-unknown-unknown")
    if err != nil || ll2 == "" { t.Fatalf("EmitModuleForTarget: %v %q", err, ll2) }
    // Write runtime and entry .ll files
    dir := filepath.Join("build", "test", "codegen_wrappers")
    _ = os.RemoveAll(dir)
    p1, err := b.WriteRuntimeLL(dir, "x86_64-unknown-linux-gnu", false)
    if err != nil { t.Fatalf("WriteRuntimeLL: %v", err) }
    if _, err := os.Stat(p1); err != nil { t.Fatalf("stat runtime: %v", err) }
    p2, err := b.WriteIngressEntrypointLL(dir, "x86_64-unknown-linux-gnu", []string{"pkg.pipe"})
    if err != nil { t.Fatalf("WriteIngressEntrypointLL: %v", err) }
    if _, err := os.Stat(p2); err != nil { t.Fatalf("stat entry: %v", err) }
    if b.TripleForEnv("darwin/amd64") == "" { t.Fatalf("TripleForEnv empty") }
}

func TestSelectDefaultBackend_SetsLLVM(t *testing.T) {
    if err := SelectDefaultBackend("llvm"); err != nil { t.Fatalf("select: %v", err) }
    if defaultBackend == nil || defaultBackend.Name() != "llvm" {
        t.Fatalf("default backend not set to llvm: %#v", defaultBackend)
    }
}

