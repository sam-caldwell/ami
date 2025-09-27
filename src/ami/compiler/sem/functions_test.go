package sem

import (
    "encoding/json"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestAnalyzeFunctions_DuplicatesAndBlanks(t *testing.T) {
    src := "package app\nfunc A(x int){}\nfunc A(y int){}\nfunc _(z int){}\nfunc B(_, w int){}\n"
    f := (&source.FileSet{}).AddFile("f.ami", src)
    p := parser.New(f)
    af, _ := p.ParseFile()
    ds := AnalyzeFunctions(af)
    if len(ds) < 3 { t.Fatalf("expected at least 3 diagnostics, got %d: %s", len(ds), codesJSON(ds)) }
    hasDup := false
    hasBlankFn := false
    hasBlankParam := false
    for _, d := range ds {
        switch d.Code {
        case "E_DUP_FUNC": hasDup = true
        case "E_BLANK_IDENT_ILLEGAL": hasBlankFn = true
        case "E_BLANK_PARAM_ILLEGAL": hasBlankParam = true
        }
    }
    if !hasDup || !hasBlankFn || !hasBlankParam {
        b, _ := json.Marshal(ds)
        t.Fatalf("missing expected codes: %s", string(b))
    }
}

