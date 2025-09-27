package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// TestParser_MapLiteral_Parsing and a simple error path to lift coverage.
func TestParser_MapLiteral_Cases(t *testing.T) {
    // Happy path
    src := "package app\nfunc F(){ m := map<int,string>{1:\"a\",2:\"b\"}; _ = m }\n"
    f := (&source.FileSet{}).AddFile("m.ami", src)
    p := New(f)
    if _, err := p.ParseFile(); err != nil { t.Fatalf("ParseFile: %v", err) }

    // Sad path: missing ':' triggers sync to next comma; remain parseable
    src2 := "package app\nfunc G(){ _ = map<int,string>{1:} }\n"
    f2 := (&source.FileSet{}).AddFile("m2.ami", src2)
    p2 := New(f2)
    if _, err := p2.ParseFile(); err != nil { t.Fatalf("ParseFile sad: %v", err) }
}
