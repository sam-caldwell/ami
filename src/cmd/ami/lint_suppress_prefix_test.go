package main

import "testing"

func Test_hasPathPrefix_basic(t *testing.T) {
    if !hasPathPrefix("/a/b", "/a") { t.Fatal("expected prefix") }
}

