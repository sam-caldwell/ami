package ascii

import (
    "testing"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

func TestFindEdge_FilePair(t *testing.T) {
    g := graph.Graph{Edges: []graph.Edge{{From:"a", To:"b"}}}
    if _, ok := findEdge(g, "a","b"); !ok { t.Fatalf("expected ok") }
}

