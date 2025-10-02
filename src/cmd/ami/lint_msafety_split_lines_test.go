package main

import "testing"

func Test_splitLinesPreserve_basic(t *testing.T) {
    got := splitLinesPreserve("a\nb")
    if len(got) != 2 { t.Fatalf("got %v", got) }
}

