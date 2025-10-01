package parser

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_PragmaCollection(t *testing.T) {
	src := "package app\n#pragma test:case name a=1 b=\"y\"\n"
	f := (&source.FileSet{}).AddFile("pr.ami", src)
	p := New(f)
	file, _ := p.ParseFile()
	if len(file.Pragmas) == 0 {
		t.Fatalf("expected pragmas collected")
	}
}
