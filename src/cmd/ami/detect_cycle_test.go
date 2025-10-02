package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

func TestDetectCycle_FilePair(t *testing.T) {
    _, _ = detectCycle(graph.Graph{})
}

