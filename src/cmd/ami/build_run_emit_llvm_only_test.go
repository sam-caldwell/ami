package main

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// When --emit-llvm-only is set, .ll is emitted but .o is skipped under build/obj/<pkg>/.
func TestRunBuild_EmitLLVMOnly_SkipsObjectCompilation(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "emit_llvm_only")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    // Write one simple unit to ensure compile sees a file
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    // Set the package-level flag to simulate --emit-llvm-only
    buildEmitLLVMOnly = true
    defer func() { buildEmitLLVMOnly = false }()
    if err := runBuild(os.Stdout, dir, false, false); err != nil { t.Fatalf("runBuild: %v", err) }
    pkg := ws.Packages[0].Package.Name
    ll := filepath.Join(dir, "build", "obj", pkg, "u.ll")
    if _, err := os.Stat(ll); err != nil { t.Fatalf("expected ll emitted: %v", err) }
    if _, err := os.Stat(filepath.Join(dir, "build", "obj", pkg, "u.o")); err == nil { t.Fatalf("did not expect .o when emit-llvm-only is set") }
}
