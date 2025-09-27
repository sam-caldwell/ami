package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestNameResolution_ImportAliasInScope(t *testing.T) {
    code := "package app\nimport alpha\nfunc F(){ alpha.X() }"
    f := &source.File{Name: "res1.ami", Content: code}
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeNameResolution(af)
    for _, d := range ds { if d.Code == "E_UNRESOLVED_IDENT" { t.Fatalf("unexpected unresolved: %v", ds) } }
}

