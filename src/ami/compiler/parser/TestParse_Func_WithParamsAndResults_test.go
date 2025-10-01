package parser

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParse_Pipeline_MissingRBrace(t *testing.T) {
	src := "package app\npipeline P(){ ingress\n"
	f := (&source.FileSet{}).AddFile("pb.ami", src)
	p := New(f)
	_, _ = p.ParseFileCollect()
}

func TestParse_Func_WithParamsAndResults(t *testing.T) {
	src := "package app\nfunc G(x int, y stringTy) (bool, rune) { return }\n"
	f := (&source.FileSet{}).AddFile("sig.ami", src)
	p := New(f)
	if _, err := p.ParseFile(); err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
}
