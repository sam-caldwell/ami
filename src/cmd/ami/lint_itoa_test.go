package main

import "testing"

func TestItoa_FilePair(t *testing.T) {
    if itoa(0) != "0" {
        t.Fatalf("itoa(0) != '0'")
    }
}

