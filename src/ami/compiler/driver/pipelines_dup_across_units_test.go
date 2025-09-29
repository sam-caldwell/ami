package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestDuplicatePipelineNamesAcrossUnits(t *testing.T) {
    fs := &source.FileSet{}
    fs.AddFile("a.ami", "package app\npipeline P(){ ingress; egress }\n")
    fs.AddFile("b.ami", "package app\npipeline P(){ ingress; egress }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    ws := workspace.Workspace{}
    _, diags := Compile(ws, pkgs, Options{})
    has := false
    for _, d := range diags { if d.Code == "E_DUP_PIPELINE" { has = true; break } }
    if !has { t.Fatalf("expected E_DUP_PIPELINE across units; got %+v", diags) }
}

