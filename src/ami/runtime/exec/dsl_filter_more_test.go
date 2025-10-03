package exec

import (
    "testing"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func Test_applyFilter_DropEven_IntAndFloat(t *testing.T) {
    e1 := ev.Event{Payload: map[string]any{"i": 2}}
    if keep := applyFilter("drop_even", e1); keep { t.Fatalf("expected drop for even int") }
    e2 := ev.Event{Payload: map[string]any{"i": 3.0}}
    if keep := applyFilter("drop_even", e2); !keep { t.Fatalf("expected keep for odd float") }
}

func Test_applyFilter_Unknown_Passes(t *testing.T) {
    e := ev.Event{Payload: map[string]any{"i": 2}}
    if !applyFilter("unknown", e) { t.Fatalf("unknown filter should pass") }
}

