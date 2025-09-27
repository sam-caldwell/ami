package main

import "testing"

func TestNewModUpdateCmd_Use(t *testing.T) {
    c := newModUpdateCmd()
    if c.Use != "update" { t.Fatalf("use: %s", c.Use) }
}

