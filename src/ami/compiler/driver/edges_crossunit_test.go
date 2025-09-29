package driver

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestCompile_Edges_CrossUnit_PipelineResolution(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // Define pipeline X with egress type Event<int> in one unit
    fs.AddFile("u1.ami", "package app\npipeline X(){ ingress; egress type(\"Event<int>\") }\n")
    // Reference X from another unit with wrong type
    fs.AddFile("u2.ami", "package app\npipeline P(){ A edge.Pipeline(name=X, type=\"Event<string>\"); egress }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(ws, pkgs, Options{Debug: false})
    var hasMismatch bool
    var hasNotFound bool
    for _, d := range diags {
        if d.Code == "E_EDGE_PIPE_TYPE_MISMATCH" { hasMismatch = true }
        if d.Code == "E_EDGE_PIPE_NOT_FOUND" { hasNotFound = true }
    }
    if !hasMismatch || hasNotFound { t.Fatalf("expect cross-unit type mismatch and no not-found; got %+v", diags) }
}

