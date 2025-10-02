package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func Test_analyzeMemorySafety_empty(t *testing.T) {
    ds := analyzeMemorySafety(&source.File{Name: "x", Content: ""})
    if len(ds) != 0 { t.Fatalf("unexpected: %+v", ds) }
}

