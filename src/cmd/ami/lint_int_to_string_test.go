package main

import "testing"

func TestIntToString_FilePair(t *testing.T) {
    if intToString(0) != "0" {
        t.Fatalf("intToString(0) != '0'")
    }
}

