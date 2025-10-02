package main

import (
    "testing"
    "time"
)

func Test_parseRate(t *testing.T) {
    if d := parseRate(""); d != 100*time.Millisecond { t.Fatalf("default: %v", d) }
    if d := parseRate("10/s"); d != 100*time.Millisecond { t.Fatalf("10/s => %v", d) }
    if d := parseRate("4/s"); d != 250*time.Millisecond { t.Fatalf("4/s => %v", d) }
    if d := parseRate("250ms"); d != 250*time.Millisecond { t.Fatalf("250ms => %v", d) }
}

