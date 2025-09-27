package main

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure we do not emit any debug artifacts when --verbose is false.
func TestRunBuild_NoVerbose_NoDebugDir(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "no_verbose_nodebug")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }

    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    if err := runBuild(os.Stdout, dir, false, false); err != nil { t.Fatalf("runBuild: %v", err) }

    // build/debug should not exist
    if _, err := os.Stat(filepath.Join(dir, "build", "debug")); err == nil || !os.IsNotExist(err) {
        t.Fatalf("expected no debug dir without --verbose; stat err=%v", err)
    }
}

