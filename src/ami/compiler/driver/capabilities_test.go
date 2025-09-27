package driver

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestCompile_Capability_IR_IOPermission(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\npipeline Z() { ingress; io.Read(\"f\"); egress }\n"
    fs.AddFile("cap.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(ws, pkgs, Options{Debug: false})
    found := false
    for _, d := range diags { if d.Code == "E_IO_PERMISSION_IR" { found = true } }
    if !found { t.Fatalf("expected E_IO_PERMISSION_IR, got %+v", diags) }
}

