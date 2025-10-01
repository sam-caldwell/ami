package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure AMI stdlib stubs for package "signal" are importable and Register can be called across packages.
func TestAMIStdlib_Signal_Stubs_Resolve(t *testing.T) {
    ws := workspace.Workspace{}
    // signal stubs are provided by builtin bundle; app imports and calls Register
    app := &source.FileSet{}
    src := "package app\nimport signal\n" +
        "func H(){}\n" +
        "func F(){ signal.Register(signal.SIGINT, H) }\n"
    app.AddFile("app.ami", src)
    pkgs := []Package{{Name: "app", Files: app}}
    _, diags := Compile(ws, pkgs, Options{Debug: false, EmitLLVMOnly: true})
    for _, d := range diags {
        if string(d.Level) == "error" {
            t.Fatalf("unexpected error diagnostic: %+v", d)
        }
    }
}

