package parser

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParse_ContainerLiteral_MissingTokens(t *testing.T) {
	// missing '>' then '{'
	src := "package app\nfunc F(){ slice<int{1}; map<int string>{1:2}; set<string>{}; }\n"
	f := (&source.FileSet{}).AddFile("cl.ami", src)
	p := New(f)
	_, _ = p.ParseFileCollect()
}
