package sem

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "testing"
)

func TestSem_FuncTypeParams_DuplicateNames_Error(t *testing.T) {
    src := `package p
func f<T any, T any>(x T) { }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics {
        if d.Code == "E_DUP_TYPE_PARAM" {
            found = true
            break
        }
    }
    if !found {
        t.Fatalf("expected E_DUP_TYPE_PARAM; got %+v", res.Diagnostics)
    }
}
