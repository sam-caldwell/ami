package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// TestParse_Pipeline_Chained_Errors exercises error in chained step name after '.'.
func TestParse_Pipeline_Chained_Errors(t *testing.T) {
    src := "package app\npipeline P(){ ingress. }\n"
    f := (&source.FileSet{}).AddFile("chainerr.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect() // tolerate parse errors; we want branch coverage
}

