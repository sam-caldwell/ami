package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestSem_Param_BlankIdentifier_Illegal(t *testing.T) {
    src := `package p
func f(_ T, y U) V { }
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_BLANK_PARAM_ILLEGAL" { found = true; break } }
    if !found { t.Fatalf("expected E_BLANK_PARAM_ILLEGAL; diags=%v", res.Diagnostics) }
}

