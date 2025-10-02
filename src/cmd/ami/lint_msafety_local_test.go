package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func Test_analyzeMemorySafety_flagsIllegal(t *testing.T) {
    f := &source.File{Name: "x.ami", Content: "&x\n* y = 1\n*bad"}
    ds := analyzeMemorySafety(f)
    if len(ds) < 2 { t.Fatalf("expected at least two diagnostics: %+v", ds) }
}

