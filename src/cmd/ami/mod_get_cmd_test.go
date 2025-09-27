package main

import "testing"

func TestNewModGetCmd_Use(t *testing.T) {
    c := newModGetCmd()
    if c.Use[:3] != "get" { t.Fatalf("use: %s", c.Use) }
}

