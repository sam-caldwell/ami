package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_MapLiteral_Errors_MissingAnglesOrBrace(t *testing.T) {
    // missing '>' after type params
    src := "package app\nfunc F(){ _ = map<int,string{1:\"a\"} }\n"
    f := (&source.FileSet{}).AddFile("mm.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect()

    // missing '{' to start map literal
    src2 := "package app\nfunc G(){ _ = map<int,string> 1 }\n"
    f2 := (&source.FileSet{}).AddFile("mm2.ami", src2)
    p2 := New(f2)
    _, _ = p2.ParseFileCollect()
}

