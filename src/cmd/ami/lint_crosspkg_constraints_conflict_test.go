package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/semver"
)

func TestConstraintsConflict_EmptyIntersection_IsConflict(t *testing.T) {
    // ^1.0.0 intersects empty with >=2.0.0
    a, err := semver.ParseConstraint("^ v1.0.0")
    if err != nil { t.Fatalf("parse a: %v", err) }
    b, err := semver.ParseConstraint(">= v2.0.0")
    if err != nil { t.Fatalf("parse b: %v", err) }
    if !constraintsConflict(a, b) {
        t.Fatalf("expected conflict between %v and %v", a, b)
    }
}

func TestConstraintsConflict_NonEmptyIntersection_IsNotConflict(t *testing.T) {
    a, err := semver.ParseConstraint(">= v1.2.0")
    if err != nil { t.Fatalf("parse a: %v", err) }
    b, err := semver.ParseConstraint("^ v1.0.0")
    if err != nil { t.Fatalf("parse b: %v", err) }
    if constraintsConflict(a, b) {
        t.Fatalf("did not expect conflict between %v and %v", a, b)
    }
}
