package main

import "testing"

func TestNewModCleanCmd_Use(t *testing.T) {
    c := newModCleanCmd()
    if c.Use != "clean" { t.Fatalf("use: %s", c.Use) }
}

