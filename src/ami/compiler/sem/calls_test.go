package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestAnalyzeCalls_ArityAndTypeMismatch(t *testing.T) {
    code := "package app\nfunc G(x int) (int){ return x }\nfunc F(){ var s string; G() }\nfunc H(){ var s string; G(s) }\n"
    f := (&source.FileSet{}).AddFile("c.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeCalls(af)
    var hasArity, hasType bool
    for _, d := range ds { if d.Code == "E_CALL_ARITY_MISMATCH" { hasArity = true }; if d.Code == "E_CALL_ARG_TYPE_MISMATCH" { hasType = true } }
    if !hasArity || !hasType { t.Fatalf("missing call diagnostics: %v", ds) }
}

