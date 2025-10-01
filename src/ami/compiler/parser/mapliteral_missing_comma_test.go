package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_MapLiteral_MissingCommaBetweenTypes(t *testing.T) {
    src := "package app\nfunc F(){ _ = map<int string>{} }\n"
    f := (&source.FileSet{}).AddFile("mmc.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect()
}

