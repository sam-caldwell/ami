package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Tolerant_Recovery_MultipleErrors(t *testing.T) {
    // Two malformed lines: bad import path and bad func header; expect both errors collected
    src := "package app\nimport 123\nfunc ( {\n}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    _, errs := p.ParseFileCollect()
    if len(errs) < 2 {
        t.Fatalf("expected at least 2 errors, got %d: %+v", len(errs), errs)
    }
}

