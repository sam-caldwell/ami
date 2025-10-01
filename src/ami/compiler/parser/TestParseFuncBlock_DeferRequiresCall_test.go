package parser

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParseFuncBlock_DeferRequiresCall(t *testing.T) {
	src := "package app\nfunc F(){ defer 1 }\n"
	f := (&source.FileSet{}).AddFile("df.ami", src)
	p := New(f)
	_, errs := p.ParseFileCollect()
	if errs == nil {
		t.Fatalf("expected errors for defer")
	}
}
