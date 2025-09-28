package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestRAII_UseAfterTransfer_NextStmt(t *testing.T) {
    code := "package app\nfunc H(a Owned){}\nfunc F(){ var a Owned; H(a); g(a) }\n"
    f := (&source.FileSet{}).AddFile("uat.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeRAII(af)
    has := false
    for _, d := range ds { if d.Code == "E_RAII_USE_AFTER_TRANSFER" { has = true } }
    if !has { t.Fatalf("expected E_RAII_USE_AFTER_TRANSFER, got: %+v", ds) }
}

func TestRAII_ReleaseAfterTransfer(t *testing.T) {
    code := "package app\nfunc H(a Owned){}\nfunc F(){ var a Owned; H(a); release(a) }\n"
    f := (&source.FileSet{}).AddFile("rat.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeRAII(af)
    has := false
    for _, d := range ds { if d.Code == "E_RAII_RELEASE_AFTER_TRANSFER" { has = true } }
    if !has { t.Fatalf("expected E_RAII_RELEASE_AFTER_TRANSFER for release after transfer, got: %+v", ds) }
}

func TestRAII_AssignAfterTransfer(t *testing.T) {
    code := "package app\nfunc H(a Owned){}\nfunc G() (Owned){ return }\nfunc F(){ var a Owned; H(a); a = G() }\n"
    f := (&source.FileSet{}).AddFile("aat.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeRAII(af)
    has := false
    for _, d := range ds { if d.Code == "E_RAII_ASSIGN_AFTER_TRANSFER" { has = true } }
    if !has { t.Fatalf("expected E_RAII_ASSIGN_AFTER_TRANSFER, got: %+v", ds) }
}
