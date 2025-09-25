package schemas

import (
    "encoding/json"
    "testing"
)

func TestPipelinesV1_ValidateAndMarshal(t *testing.T) {
    p := &PipelinesV1{Schema: "pipelines.v1", Package: "p", File: "f.ami", Pipelines: []PipelineV1{{Name: "P"}}}
    if err := p.Validate(); err != nil { t.Fatalf("validate failed: %v", err) }
    b, err := json.Marshal(p)
    if err != nil || len(b) == 0 { t.Fatalf("marshal failed: %v", err) }
    var got PipelinesV1
    if err := json.Unmarshal(b, &got); err != nil { t.Fatalf("unmarshal failed: %v", err) }
    if got.Schema != "pipelines.v1" || got.Package != "p" || got.File != "f.ami" { t.Fatalf("round-trip mismatch: %+v", got) }
}

func TestPipelinesV1_Validate_SadSchema(t *testing.T) {
    p := &PipelinesV1{Schema: "wrong"}
    if err := p.Validate(); err == nil { t.Fatalf("expected error for wrong schema") }
}

