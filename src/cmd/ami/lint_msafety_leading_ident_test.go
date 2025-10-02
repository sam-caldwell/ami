package main

import "testing"

func Test_leadingIdent_basic(t *testing.T) {
    if leadingIdent("abc=1") != "abc" { t.Fatal("bad") }
}

