package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestReturnTypes_Match(t *testing.T) {
    code := "package app\nfunc F() (int) { return 1 }\n"
    f := (&source.FileSet{}).AddFile("t.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeReturnTypes(af)
    if len(ds) != 0 { t.Fatalf("expected no diags, got %d", len(ds)) }
}

func TestReturnTypes_Mismatch_Type(t *testing.T) {
    code := "package app\nfunc F() (string) { return 1 }\n"
    f := (&source.FileSet{}).AddFile("t.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeReturnTypes(af)
    if len(ds) == 0 { t.Fatalf("expected mismatch") }
}

func TestReturnTypes_Mismatch_Arity(t *testing.T) {
    code := "package app\nfunc F() (int, string) { return 1 }\n"
    f := (&source.FileSet{}).AddFile("t.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeReturnTypes(af)
    if len(ds) == 0 { t.Fatalf("expected arity mismatch") }
}

