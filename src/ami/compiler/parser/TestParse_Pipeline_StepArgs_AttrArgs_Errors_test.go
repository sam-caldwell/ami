package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// TestParse_Pipeline_StepArgs_AttrArgs_Errors hits unexpected token branches in args and attr args.
func TestParse_Pipeline_StepArgs_AttrArgs_Errors(t *testing.T) {
    src := "package app\npipeline P(){ Alpha(,).Collect edge.MultiPath(,) ; egress }\n"
    f := (&source.FileSet{}).AddFile("argsattr.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect()
}

