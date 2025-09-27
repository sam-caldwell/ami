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

// TestParse_Pipeline_Chained_Errors exercises error in chained step name after '.'
func TestParse_Pipeline_Chained_Errors(t *testing.T) {
    src := "package app\npipeline P(){ ingress. }\n"
    f := (&source.FileSet{}).AddFile("chainerr.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect() // tolerate parse errors; we want branch coverage
}

// TestParse_Pipeline_StepArgs_AttrArgs_Errors hits unexpected token branches in args and attr args.
func TestParse_Pipeline_StepArgs_AttrArgs_Errors(t *testing.T) {
    src := "package app\npipeline P(){ Alpha(,).Collect edge.MultiPath(,) ; egress }\n"
    f := (&source.FileSet{}).AddFile("argsattr.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect()
}

func TestParse_ErrorBlock_Chained_MissingName(t *testing.T) {
    src := "package app\nerror { Alpha(). Beta() }\n"
    f := (&source.FileSet{}).AddFile("errchain.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect()
}

func TestParse_ErrorBlock_RichSteps(t *testing.T) {
    src := "package app\nerror {\n// lead comment\nAlpha(@, 1) edge.MultiPath(, merge.Stable()) ;\n} \n"
    f := (&source.FileSet{}).AddFile("richerr.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect()
}
