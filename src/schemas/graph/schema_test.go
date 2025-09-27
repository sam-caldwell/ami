package graph

import "testing"

func TestSchema_Constant(t *testing.T) {
    if Schema != "graph.v1" { t.Fatalf("schema: %s", Schema) }
}

