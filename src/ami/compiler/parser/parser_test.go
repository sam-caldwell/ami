package parser

import "testing"

func TestExtractImports_SingleAndBlock(t *testing.T) {
    src := `package main
import "a/one"
import (
    "b/two"
    alias "c/three"
)
`
    got := ExtractImports(src)
    want := []string{"a/one", "b/two", "c/three"}
    if len(got) != len(want) {
        t.Fatalf("imports length mismatch: got %d want %d", len(got), len(want))
    }
    for i := range want {
        if got[i] != want[i] {
            t.Fatalf("imports[%d]=%q want %q", i, got[i], want[i])
        }
    }
}

func TestParseFile_PackageAndImports(t *testing.T) {
    src := `package pkg
import ("x/y")`
    p := New(src)
    f := p.ParseFile()
    if f.Package != "pkg" { t.Fatalf("package=%q want pkg", f.Package) }
    if len(f.Imports) != 1 || f.Imports[0] != "x/y" {
        t.Fatalf("imports=%v want [x/y]", f.Imports)
    }
}

