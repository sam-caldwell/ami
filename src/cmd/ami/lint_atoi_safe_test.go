package main

import "testing"

func TestAtoiSafe_FilePair(t *testing.T) {
    if atoiSafe("123") != 123 {
        t.Fatalf("atoiSafe failed")
    }
}

