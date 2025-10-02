package main

import (
    "testing"
    ast "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestCallKey_FilePair(t *testing.T) {
    if got := callKey("", (*ast.CallExpr)(nil)); got != "" {
        t.Fatalf("unexpected: %q", got)
    }
}

