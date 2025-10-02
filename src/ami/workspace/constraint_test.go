package workspace

import "testing"

func TestConstraint_ExactSatisfies(t *testing.T) {
    c, err := ParseConstraint("v1.2.3")
    if err != nil { t.Fatalf("parse: %v", err) }
    if !Satisfies("1.2.3", c) { t.Fatalf("expected satisfy") }
}

