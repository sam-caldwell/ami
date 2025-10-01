package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParse_ErrorBlock_Chained_MissingName(t *testing.T) {
    src := "package app\nerror { Alpha(). Beta() }\n"
    f := (&source.FileSet{}).AddFile("errchain.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect()
}

