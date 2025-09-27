package workspace

import (
    "os"
    "path/filepath"
    "testing"
)

func TestWorkspace_Load_DeduplicatesEnvPreservingOrder(t *testing.T) {
    dir := filepath.Join("build", "test", "workspace_env", "dedup")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    path := filepath.Join(dir, "ami.workspace")
    y := []byte("version: 1.0.0\ntoolchain:\n  compiler:\n    env: [linux/amd64, linux/arm64, linux/amd64, darwin/arm64]\npackages: []\n")
    if err := os.WriteFile(path, y, 0o644); err != nil { t.Fatalf("write: %v", err) }
    var w Workspace
    if err := w.Load(path); err != nil { t.Fatalf("load: %v", err) }
    got := w.Toolchain.Compiler.Env
    want := []string{"linux/amd64", "linux/arm64", "darwin/arm64"}
    if len(got) != len(want) { t.Fatalf("len=%d want %d: %v", len(got), len(want), got) }
    for i := range want {
        if got[i] != want[i] { t.Fatalf("[%d]=%q want %q; %v", i, got[i], want[i], got) }
    }
}

