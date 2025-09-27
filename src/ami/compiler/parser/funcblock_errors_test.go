package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParseFuncBlock_DeferRequiresCall(t *testing.T) {
    src := "package app\nfunc F(){ defer 1 }\n"
    f := (&source.FileSet{}).AddFile("df.ami", src)
    p := New(f)
    _, errs := p.ParseFileCollect()
    if errs == nil { t.Fatalf("expected errors for defer") }
}

func TestParseFuncBlock_StarAssignErrors(t *testing.T) {
    src := "package app\nfunc F(){ * 1 = 2; *x 1 }\n"
    f := (&source.FileSet{}).AddFile("sa.ami", src)
    p := New(f)
    _, errs := p.ParseFileCollect()
    if errs == nil { t.Fatalf("expected errors for star assign") }
}

func TestParse_ContainerLiteral_MissingTokens(t *testing.T) {
    // missing '>' then '{'
    src := "package app\nfunc F(){ slice<int{1}; map<int string>{1:2}; set<string>{}; }\n"
    f := (&source.FileSet{}).AddFile("cl.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect()
}

func TestParse_Pipeline_MissingRBrace(t *testing.T) {
    src := "package app\npipeline P(){ ingress\n"
    f := (&source.FileSet{}).AddFile("pb.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect()
}

func TestParse_FuncBlock_UnknownToken(t *testing.T) {
    src := "package app\nfunc F(){ @ }\n"
    f := (&source.FileSet{}).AddFile("utok.ami", src)
    p := New(f)
    _, _ = p.ParseFileCollect()
}
