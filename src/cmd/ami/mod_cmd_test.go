package main

import "testing"

func TestNewModCmd_Use(t *testing.T) {
    c := newModCmd()
    if c.Use != "mod" { t.Fatalf("use: %s", c.Use) }
    if len(c.Commands()) == 0 { t.Fatalf("expected subcommands") }
}

