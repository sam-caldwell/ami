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
