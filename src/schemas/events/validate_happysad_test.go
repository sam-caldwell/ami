package events

import "testing"

func TestValidate_HappyAndSadPaths(t *testing.T) {
    good := Event{ID: "x"}
    if err := Validate(good); err != nil { t.Fatalf("good event invalid: %v", err) }
    bad := Event{}
    if err := Validate(bad); err == nil { t.Fatalf("expected error for missing id") }
}

