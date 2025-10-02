package main

import "testing"

func Test_parseKV_basic(t *testing.T) {
    m := parseKV("a=1 b='two' c=\"3\"")
    if m["a"] != "1" || m["b"] != "two" || m["c"] != "3" { t.Fatalf("got: %#v", m) }
}

