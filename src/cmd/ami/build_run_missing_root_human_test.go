package main

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/exit"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_Human_MissingPackageRoot_IOExit(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "human_missing_root")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./does_not_exist"
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    err := runBuild(os.Stdout, dir, false, true)
    if err == nil { t.Fatalf("expected error") }
    if exit.UnwrapCode(err) != exit.IO { t.Fatalf("want IO exit, got %v", exit.UnwrapCode(err)) }
}

