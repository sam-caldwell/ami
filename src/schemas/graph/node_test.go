package graph

import "testing"

func TestNode_BasenamePair(t *testing.T) {
    n := Node{ID: "n1", Kind: "ingress", Label: "A"}
    if n.ID != "n1" || n.Kind != "ingress" { t.Fatalf("unexpected: %+v", n) }
}

