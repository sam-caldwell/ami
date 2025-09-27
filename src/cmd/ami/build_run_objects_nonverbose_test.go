package main

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_NonVerbose_EmitsObjectIndex(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "nonverbose_objects")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := runBuild(os.Stdout, dir, false, false); err != nil { t.Fatalf("runBuild: %v", err) }
    // object index exists
    p := filepath.Join(dir, "build", "obj", ws.Packages[0].Package.Name, "index.json")
    if _, err := os.Stat(p); err != nil { t.Fatalf("obj index missing: %v", err) }
    // ensure no debug dir created
    if _, err := os.Stat(filepath.Join(dir, "build", "debug")); err == nil { t.Fatalf("debug dir should not exist") }
}

