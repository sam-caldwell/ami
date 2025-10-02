package ascii

import (
    "testing"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

func TestRenderLine_FilePair(t *testing.T) {
    g := graph.Graph{Nodes: []graph.Node{{ID:"a", Kind:"ingress", Label:"A"}, {ID:"b", Kind:"egress", Label:"B"}}, Edges: []graph.Edge{{From:"a", To:"b"}}}
    got := RenderLine(g, Options{})
    if got == "" { t.Fatalf("expected non-empty") }
}

