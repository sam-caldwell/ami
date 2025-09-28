package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestInference_Modulo_Int(t *testing.T) {
    src := "package app\nfunc F(){ var x int; x = 5 % 2 }\n"
    f := (&source.FileSet{}).AddFile("mod.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeTypeInference(af)
    if len(ds) != 0 { t.Fatalf("unexpected diags: %+v", ds) }
}

func TestInference_Modulo_Mismatch(t *testing.T) {
    src := "package app\nfunc F(){ var s string; s = 5 % 2 }\n"
    f := (&source.FileSet{}).AddFile("mod_bad.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeTypeInference(af)
    has := false
    for _, d := range ds { if d.Code == "E_TYPE_MISMATCH" { has = true } }
    if !has { t.Fatalf("expected mismatch for modulo to string: %+v", ds) }
}

