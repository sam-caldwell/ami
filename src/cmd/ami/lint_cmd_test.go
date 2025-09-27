package main

import "testing"

func TestNewLintCmd_Use(t *testing.T) {
    c := newLintCmd()
    if c.Use != "lint" { t.Fatalf("use: %s", c.Use) }
}

