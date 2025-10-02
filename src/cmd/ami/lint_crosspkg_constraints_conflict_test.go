package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/semver"
)

func Test_constraintsConflict_trivial(t *testing.T) {
    a, _ := semver.ParseConstraint(">=1.0.0")
    b, _ := semver.ParseConstraint("<2.0.0")
    if constraintsConflict(a, b) { t.Fatal("should not conflict") }
}

