package main

import "testing"

func Test_strOrEmpty(t *testing.T) {
    if got := strOrEmpty("x"); got != "x" { t.Fatalf("got %q", got) }
    if got := strOrEmpty(123); got != "" { t.Fatalf("expected empty, got %q", got) }
}

