package ascii

import (
    "testing"
    "github.com/sam-caldwell/ami/src/schemas/graph"
)

func TestEdgeTag_FilePair(t *testing.T) {
    e := graph.Edge{Attrs: map[string]any{"bounded":true,"delivery":"bestEffort","type":"X"}}
    if edgeTag(e) == "" { t.Fatalf("expected non-empty tag") }
}

