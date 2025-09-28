package main

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// For linux/arm64 env, ensure per-env .ll (and optionally .o) emission occurs and index is written when possible.
func TestRunBuild_LinuxArm64_EnvArtifacts(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "linux_arm64_artifacts")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Toolchain.Compiler.Env = []string{"linux/arm64"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    if err := runBuild(os.Stdout, dir, false, false); err != nil { t.Fatalf("runBuild: %v", err) }

    // Expect per-env LLVM IR emitted under build/linux/arm64/obj/<pkg>/u.ll
    objDir := filepath.Join(dir, "build", "linux", "arm64", "obj", ws.Packages[0].Package.Name)
    if st, err := os.Stat(objDir); err != nil || !st.IsDir() { t.Fatalf("env obj dir missing: %v", err) }
    // There should be a .ll file
    matches, _ := filepath.Glob(filepath.Join(objDir, "*.ll"))
    if len(matches) == 0 { t.Fatalf("expected per-env .ll under %s", objDir) }
}

