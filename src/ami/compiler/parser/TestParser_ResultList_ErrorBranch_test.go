package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_ResultList_ErrorBranch(t *testing.T) {
    src := "package app\nfunc X() (int, ) { return }\n"
    f := (&source.FileSet{}).AddFile("res.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect()
}

