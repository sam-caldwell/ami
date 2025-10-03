package main

import "testing"

func TestValidPkgName_PositiveCases(t *testing.T) {
    good := []string{"alpha", "alphaBeta", "AlphaBeta", "x1", "Z9"}
    for _, s := range good {
        if !validPkgName(s) { t.Fatalf("expected valid: %q", s) }
    }
}

