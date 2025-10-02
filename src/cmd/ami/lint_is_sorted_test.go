package main

import "testing"

func TestIsSorted_FilePair(t *testing.T) {
    if !isSorted([]string{"a","b"}) { t.Fatalf("expected sorted") }
}

