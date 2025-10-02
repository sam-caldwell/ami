package main

import "testing"

func Test_selectHighestSemver_basic(t *testing.T) {
    v, err := selectHighestSemver([]string{"v0.9.0", "v1.2.3", "v1.2.3-beta.1"}, false)
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if v != "v1.2.3" { t.Fatalf("got %q", v) }
}

