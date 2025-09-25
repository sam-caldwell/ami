package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestSem_BlankImportAlias_Illegal(t *testing.T) {
    src := `package p
import _ "x/y"
`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    var seen bool
    for _, d := range res.Diagnostics { if d.Code == "E_BLANK_IMPORT_ALIAS" { seen = true; break } }
    if !seen { t.Fatalf("expected E_BLANK_IMPORT_ALIAS; got %+v", res.Diagnostics) }
}

func TestSem_BlankFunc_Illegal(t *testing.T) {
    src := `package p
func _() {}`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    var seen bool
    for _, d := range res.Diagnostics { if d.Code == "E_BLANK_IDENT_ILLEGAL" { seen = true; break } }
    if !seen { t.Fatalf("expected E_BLANK_IDENT_ILLEGAL; got %+v", res.Diagnostics) }
}
