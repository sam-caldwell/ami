package ascii

import (
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/schemas/graph"
)

func TestRenderBlock_InlineEdgeTags(t *testing.T) {
    g := graph.Graph{
        Package: "main",
        Name:    "Tags",
        Nodes: []graph.Node{
            {ID: "00:ingress", Kind: "ingress", Label: "ingress"},
            {ID: "01:worker",  Kind: "worker",  Label: "worker"},
        },
        Edges: []graph.Edge{
            {From: "00:ingress", To: "01:worker", Attrs: map[string]any{"bounded": true, "delivery": "bestEffort", "type": "X"}},
        },
    }
    s := RenderBlock(g, Options{})
    if !strings.Contains(s, "--[bounded,bestEffort,type:X]-->") {
        t.Fatalf("missing inline tags: %q", s)
    }
}

func TestRenderBlock_FocusHighlight(t *testing.T) {
    g := graph.Graph{
        Package: "main",
        Name:    "Focus",
        Nodes: []graph.Node{
            {ID: "00:ingress", Kind: "ingress", Label: "ingress"},
            {ID: "01:worker",  Kind: "worker",  Label: "worker"},
        },
        Edges: []graph.Edge{{From: "00:ingress", To: "01:worker"}},
    }
    s := RenderBlock(g, Options{Focus: "work"})
    if !strings.Contains(s, "(*worker*)") {
        t.Fatalf("missing focus highlighting: %q", s)
    }
}

