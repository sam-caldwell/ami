package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// TestParse_StepBlock_UnknownToken exercises the recovery path in parseStepBlock.
func TestParse_StepBlock_UnknownToken(t *testing.T) {
    src := "package app\npipeline P(){ : ; egress }\n"
    f := (&source.FileSet{}).AddFile("sb.ami", src)
    p := New(f)
    if _, errs := p.ParseFileCollect(); errs == nil {
        t.Fatalf("expected errors, got nil")
    }
}

