package main

import "testing"

func TestContainsUnderscore_FilePair(t *testing.T) {
    if !containsUnderscore("a_b") {
        t.Fatalf("expected true")
    }
}

