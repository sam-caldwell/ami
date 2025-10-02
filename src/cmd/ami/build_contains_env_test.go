package main

import "testing"

func TestContainsEnv_FilePair(t *testing.T) {
    if !containsEnv([]string{"a","b"}, "a") {
        t.Fatalf("containsEnv failed")
    }
}

