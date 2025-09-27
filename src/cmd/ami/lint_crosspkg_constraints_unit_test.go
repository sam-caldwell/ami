package main

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/semver"
)

func TestConstraintsConflict_Helper(t *testing.T) {
    // Non-overlapping ranges: >= v2.0.0 and < v2.0.0
    a, err := semver.ParseConstraint(">= v2.0.0")
    if err != nil { t.Fatalf("parse: %v", err) }
    b, err := semver.ParseConstraint("< v2.0.0")
    if err != nil { t.Fatalf("parse: %v", err) }
    if !constraintsConflict(a, b) {
        t.Fatalf("expected conflict between %v and %v", a, b)
    }
    // Overlapping: >= v1.0.0 and < v3.0.0
    c, _ := semver.ParseConstraint(">= v1.0.0")
    d, _ := semver.ParseConstraint("< v3.0.0")
    if constraintsConflict(c, d) {
        t.Fatalf("did not expect conflict between %v and %v", c, d)
    }
}

