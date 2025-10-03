package exec

import (
    "testing"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func Test_applyTransform_AddField(t *testing.T) {
    e := ev.Event{Payload: map[string]any{"x": 1}}
    out := applyTransform("add_field:newflag", e)
    m := out.Payload.(map[string]any)
    if m["newflag"] != true { t.Fatalf("expected newflag=true, got: %v", m["newflag"]) }
}

