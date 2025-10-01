package parser

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParse_FuncBlock_UnknownToken(t *testing.T) {
	src := "package app\nfunc F(){ @ }\n"
	f := (&source.FileSet{}).AddFile("utok.ami", src)
	p := New(f)
	_, _ = p.ParseFileCollect()
}
