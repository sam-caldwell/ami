package main

import "testing"

func Test_key(t *testing.T) {
    if got := key("pkg", ""); got != "pkg" { t.Fatalf("got %q", got) }
    if got := key("pkg", "1.2.3"); got != "pkg@1.2.3" { t.Fatalf("got %q", got) }
}

