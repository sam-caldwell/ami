package ascii

import (
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/schemas/graph"
)

func TestRenderBlock_SimpleBranch(t *testing.T) {
    g := graph.Graph{
        Package: "main",
        Name:    "Branch",
        Nodes: []graph.Node{
            {ID: "00:ingress", Kind: "ingress", Label: "ingress"},
            {ID: "01:worker", Kind: "worker", Label: "worker"},
            {ID: "02:egress", Kind: "egress", Label: "egress"},
            {ID: "03:err",    Kind: "worker", Label: "errorhandler"},
        },
        Edges: []graph.Edge{
            {From: "00:ingress", To: "01:worker"},
            {From: "01:worker",  To: "02:egress"},
            {From: "01:worker",  To: "03:err"},
        },
    }
    s := RenderBlock(g, Options{})
    if !strings.Contains(s, "[ingress] --> (worker) --> [egress]\n") {
        t.Fatalf("missing main chain: %q", s)
    }
    if !strings.Contains(s, "|\n") || !strings.Contains(s, "+--> (errorhandler)\n") {
        t.Fatalf("missing branch lines: %q", s)
    }
}

