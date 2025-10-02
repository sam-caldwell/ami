package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

func TestApplyJSONExcludes_FilePair(t *testing.T) {
    _ = applyJSONExcludes(graph.Graph{}, []string{"attrs"})
}

