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

func TestReturnTypesWithSigs_TupleReturn_FromSingleCall_OK(t *testing.T) {
    code := "package app\nfunc Pair() (int,string) { return 1,\"x\" }\nfunc F() (int,string) { return Pair() }\n"
    f := (&source.FileSet{}).AddFile("t.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    res := map[string][]string{"Pair": {"int", "string"}}
    ds := AnalyzeReturnTypesWithSigs(af, res)
    if len(ds) != 0 { t.Fatalf("unexpected diags: %+v", ds) }
}

func TestReturnTypesWithSigs_LocalEnv_InferredIdent_OK(t *testing.T) {
    code := "package app\nfunc F() (string) { var a string; a = \"x\"; return a }\n"
    f := (&source.FileSet{}).AddFile("t2.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeReturnTypesWithSigs(af, map[string][]string{})
    if len(ds) != 0 { t.Fatalf("unexpected diags: %+v", ds) }
}

func TestReturnTypesWithSigs_LocalEnv_InferredIdent_Mismatch(t *testing.T) {
    code := "package app\nfunc F() (int) { var a string; a = \"x\"; return a }\n"
    f := (&source.FileSet{}).AddFile("t3.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeReturnTypesWithSigs(af, map[string][]string{})
    if len(ds) == 0 { t.Fatalf("expected mismatch diag") }
}
