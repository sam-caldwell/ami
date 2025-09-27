package graph

import "testing"

func TestEdge_BasenamePair(t *testing.T) {
    e := Edge{From: "a", To: "b"}
    if e.From != "a" || e.To != "b" { t.Fatalf("unexpected: %+v", e) }
}

