package main

import "testing"

func TestSplitImportConstraint_FilePair(t *testing.T) {
    p, c := splitImportConstraint("a/b >= v1.2.3")
    if p == "" && c == "" { t.Fatalf("unexpected empty result") }
}

