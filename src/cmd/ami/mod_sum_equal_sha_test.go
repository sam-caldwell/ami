package main

import "testing"

func Test_equalSHA(t *testing.T) {
    if !equalSHA("abc", "abc") { t.Fatal("expected equal") }
    if equalSHA("abc", "abcd") { t.Fatal("length mismatch should be false") }
}

