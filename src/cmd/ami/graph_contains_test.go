package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

func TestGraphContains_FilePair(t *testing.T) {
    g := graph.Graph{Name: "example"}
    _ = graphContains(g, "ex")
}

