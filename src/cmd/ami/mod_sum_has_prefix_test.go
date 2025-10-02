package main

import "testing"

func Test_hasPrefix(t *testing.T) {
    if !hasPrefix("abcdef", "abc") { t.Fatal("expected true") }
    if hasPrefix("ab", "abc") { t.Fatal("expected false") }
}

