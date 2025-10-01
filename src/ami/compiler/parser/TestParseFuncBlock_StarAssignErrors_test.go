package parser

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParseFuncBlock_StarAssignErrors(t *testing.T) {
	src := "package app\nfunc F(){ * 1 = 2; *x 1 }\n"
	f := (&source.FileSet{}).AddFile("sa.ami", src)
	p := New(f)
	_, errs := p.ParseFileCollect()
	if errs == nil {
		t.Fatalf("expected errors for star assign")
	}
}
