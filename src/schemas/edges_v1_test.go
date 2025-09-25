package schemas

import (
    "encoding/json"
    "testing"
)

func TestEdgesV1_ValidateAndMarshal(t *testing.T) {
    e := &EdgesV1{Schema: "edges.v1", Package: "p", Items: []EdgeInitV1{{Pipeline:"P", Segment:"normal", Step:0, Node:"Egress", Label:"P.step0.in", Kind:"edge.FIFO"}}}
    if err := e.Validate(); err != nil { t.Fatalf("validate failed: %v", err) }
    b, err := json.Marshal(e)
    if err != nil || len(b) == 0 { t.Fatalf("marshal failed: %v", err) }
    var got EdgesV1
    if err := json.Unmarshal(b, &got); err != nil { t.Fatalf("unmarshal: %v", err) }
    if got.Schema != "edges.v1" || got.Package != "p" { t.Fatalf("round-trip mismatch: %+v", got) }
}

func TestEdgesV1_Validate_SadSchema(t *testing.T) {
    e := &EdgesV1{Schema: "wrong"}
    if err := e.Validate(); err == nil { t.Fatalf("expected schema error") }
}

