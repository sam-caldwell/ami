package driver

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestCompile_MergeField_Struct_ResolveAndOrderable(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // Upstream A declares Event<Struct{id:int,ts:int}>; Sort(ts) is valid
    fs.AddFile("u.ami", "package app\npipeline P(){ ingress; transform type(\"Event<Struct{id:int,ts:int}>\"); ingress -> transform; transform -> Collect; Collect merge.Sort(\"ts\"); Collect -> egress; egress }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, ds := Compile(ws, pkgs, Options{Debug: false})
    for _, d := range ds { if d.Code == "E_MERGE_SORT_FIELD_UNKNOWN" || d.Code == "E_MERGE_SORT_FIELD_UNORDERABLE" { t.Fatalf("unexpected diag: %+v", ds) } }
}

func TestCompile_MergeField_Struct_Unknown(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("u.ami", "package app\npipeline P(){ ingress; transform type(\"Event<Struct{id:int}>\"); ingress -> transform; transform -> Collect; Collect merge.Sort(\"missing\"); Collect -> egress; egress }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, ds := Compile(ws, pkgs, Options{Debug: false})
    has := false
    for _, d := range ds { if d.Code == "E_MERGE_SORT_FIELD_UNKNOWN" { has = true } }
    if !has { t.Fatalf("expected E_MERGE_SORT_FIELD_UNKNOWN; got %+v", ds) }
}

func TestCompile_MergeField_Struct_Unorderable(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // Field obj is a nested Struct; not orderable
    fs.AddFile("u.ami", "package app\npipeline P(){ ingress; transform type(\"Event<Struct{obj:Struct{x:int}}>\"); ingress -> transform; transform -> Collect; Collect merge.Sort(\"obj\"); Collect -> egress; egress }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, ds := Compile(ws, pkgs, Options{Debug: false})
    has := false
    for _, d := range ds { if d.Code == "E_MERGE_SORT_FIELD_UNORDERABLE" { has = true } }
    if !has { t.Fatalf("expected E_MERGE_SORT_FIELD_UNORDERABLE; got %+v", ds) }
}
