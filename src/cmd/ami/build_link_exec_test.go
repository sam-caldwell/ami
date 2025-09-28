package main

import (
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "testing"

    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// End-to-end: ami build produces a runnable binary when clang is available.
func TestRunBuild_LinksExecutable_WhenToolchainPresent(t *testing.T) {
    if _, err := llvme.FindClang(); err != nil { t.Skip("clang not found") }
    dir := filepath.Join("build", "test", "ami_build", "link_exec")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save ws: %v", err) }
    // minimal AMI source for the main package
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package newProject\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := runBuild(os.Stdout, dir, false, false); err != nil { t.Fatalf("runBuild: %v", err) }
    env := runtime.GOOS + "/" + runtime.GOARCH
    bin := filepath.Join(dir, "build", env, "newProject")
    if runtime.GOOS == "windows" { bin += ".exe" }
    if st, err := os.Stat(bin); err != nil || st.IsDir() { t.Fatalf("binary missing: %v st=%v", err, st) }
    // execute binary; should exit 0
    cmd := exec.Command(bin)
    if err := cmd.Run(); err != nil { t.Fatalf("run bin: %v", err) }
}
