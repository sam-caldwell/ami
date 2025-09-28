package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

func TestFindPackageByRootKey(t *testing.T) {
    ws := workspace.Workspace{
        Packages: []workspace.PackageEntry{
            {Key: "main", Package: workspace.Package{Name: "app", Root: "./src"}},
            {Key: "util", Package: workspace.Package{Name: "util", Root: "./util"}},
        },
    }
    if p := findPackageByRootKey(&ws, "main"); p == nil || p.Name != "app" {
        t.Fatalf("expected to find main/app, got %+v", p)
    }
    if p := findPackageByRootKey(&ws, "missing"); p != nil {
        t.Fatalf("expected nil for missing key, got %+v", p)
    }
}

func TestGraphContains(t *testing.T) {
    g := graph.Graph{
        Package: "pkg", Name: "PipeLineA",
        Nodes: []graph.Node{{ID: "n1", Kind: "ingress", Label: "source"}, {ID: "n2", Kind: "egress", Label: "sink"}},
    }
    if !graphContains(g, "pipe") { t.Fatalf("expected match on pipeline name") }
    if !graphContains(g, "sink") { t.Fatalf("expected match on node label") }
    if graphContains(g, "nomatch") { t.Fatalf("did not expect match") }
}

