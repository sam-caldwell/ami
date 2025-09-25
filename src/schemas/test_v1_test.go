package schemas

import (
    "encoding/json"
    "testing"
)

func TestTestV1_ValidateAndMarshal(t *testing.T) {
    ev := &TestRunStart{Schema:"test.v1", Type:"run_start", Workspace:".", Packages: []string{"pkg/a"}}
    if err := ev.Validate(); err != nil { t.Fatalf("validate: %v", err) }
    b, err := json.Marshal(ev)
    if err != nil { t.Fatalf("marshal: %v", err) }
    var got TestRunStart
    if err := json.Unmarshal(b, &got); err != nil { t.Fatalf("unmarshal: %v", err) }
    if got.Schema != "test.v1" || got.Type != "run_start" { t.Fatalf("unexpected: %+v", got) }
}

