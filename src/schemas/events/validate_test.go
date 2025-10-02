package events

import "testing"

func TestValidate_RequiresID(t *testing.T) {
    if err := Validate(Event{}); err == nil {
        t.Fatalf("expected error for missing id")
    }
}

