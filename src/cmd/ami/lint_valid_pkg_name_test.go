package main

import "testing"

func TestValidPkgName_FilePair(t *testing.T) {
    if validPkgName("bad_name") {
        t.Fatalf("expected false")
    }
}

