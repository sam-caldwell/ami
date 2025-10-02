package main

import "testing"

func TestStringsHasPrefixAny_FilePair(t *testing.T) {
    if !stringsHasPrefixAny("./x", []string{"./"}) {
        t.Fatalf("expected true")
    }
}

