package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/semver"
)

func Test_constraintsConflict_conflictingRanges(t *testing.T) {
    // Use only supported ops in semver.ParseConstraint: ^, ~, >=, >, or exact
    a, _ := semver.ParseConstraint(">=2.0.0")
    b, _ := semver.ParseConstraint("1.0.0") // exact 1.0.0
    if !constraintsConflict(a, b) {
        t.Fatal("expected constraints to conflict")
    }
}
