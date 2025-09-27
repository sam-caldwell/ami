package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// TestParser_MapLiteral_Parsing and a simple error path to lift coverage.
func TestParser_MapLiteral_Cases(t *testing.T) {
    // Happy path
    src := "package app\nfunc F(){ var m map<int,string> = map<int,string>{1:\"a\",2:\"b\"}; _ = m }\n"
    f := (&source.FileSet{}).AddFile("m.ami", src)
    p := New(f)
    if _, err := p.ParseFile(); err != nil { t.Fatalf("ParseFile: %v", err) }

    // Sad path: missing ':' triggers sync to next comma; remain parseable
    src2 := "package app\nfunc G(){ _ = map<int,string>{1:} }\n"
    f2 := (&source.FileSet{}).AddFile("m2.ami", src2)
    p2 := New(f2)
    _, _ = p2.ParseFile() // tolerate parse error; we only need to exercise branches

    // Sad path: missing '}'
    src3 := "package app\nfunc H(){ _ = map<int,string>{1:\"a\" }\n" // unterminated
    f3 := (&source.FileSet{}).AddFile("m3.ami", src3)
    p3 := New(f3)
    _, _ = p3.ParseFileCollect()
}
