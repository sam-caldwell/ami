package main

import "testing"

func TestNewModListCmd_Use(t *testing.T) {
    c := newModListCmd()
    if c.Use != "list" { t.Fatalf("use: %s", c.Use) }
}

