package main

import (
    "bytes"
    "os"
    "path/filepath"
    "runtime"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/exit"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_WorkspaceSchemaError_Human(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "schema_human")
    if err := os.MkdirAll(dir, 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    // Write an invalid workspace: absolute target
    ws := workspace.DefaultWorkspace()
    if runtime.GOOS == "windows" {
        ws.Toolchain.Compiler.Target = filepath.Join(string(filepath.Separator), "abs")
    } else {
        ws.Toolchain.Compiler.Target = "/abs"
    }
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
        t.Fatalf("save: %v", err)
    }
    var buf bytes.Buffer
    err := runBuild(&buf, dir, false, false)
    if err == nil { t.Fatalf("expected error") }
    if exit.UnwrapCode(err) != exit.User {
        t.Fatalf("expected User exit; got %v", exit.UnwrapCode(err))
    }
    // Human mode writes error via root execution; runBuild returns error here without writing to out.
}
