package ascii

import (
    "testing"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

func TestRenderBlock_FilePair(t *testing.T) {
    g := graph.Graph{Nodes: []graph.Node{{ID:"a", Kind:"ingress", Label:"A"}, {ID:"b", Kind:"egress", Label:"B"}}, Edges: []graph.Edge{{From:"a", To:"b"}}}
    got := RenderBlock(g, Options{Legend: true})
    if got == "" { t.Fatalf("expected non-empty") }
}

