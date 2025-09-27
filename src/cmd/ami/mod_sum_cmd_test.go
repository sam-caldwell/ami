package main

import "testing"

func TestNewModSumCmd_Use(t *testing.T) {
    c := newModSumCmd()
    if c.Use != "sum" { t.Fatalf("use: %s", c.Use) }
}

