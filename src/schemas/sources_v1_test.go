package schemas

import (
    "encoding/json"
    "testing"
)

func TestSourcesV1_ValidateAndMarshal(t *testing.T) {
    s := &SourcesV1{Schema:"sources.v1", Units: []SourceUnit{{Package:"p", File:"f", Imports:[]string{"x"}, Source:"code"}}}
    if err := s.Validate(); err != nil { t.Fatalf("validate: %v", err) }
    data, err := json.Marshal(s)
    if err != nil { t.Fatalf("marshal: %v", err) }
    var got SourcesV1
    if err := json.Unmarshal(data, &got); err != nil { t.Fatalf("unmarshal: %v", err) }
    if got.Schema != "sources.v1" { t.Fatalf("unexpected schema: %s", got.Schema) }
}

