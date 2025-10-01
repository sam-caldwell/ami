package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// Ensure E_GENERIC_ARITY_MISMATCH on returns includes expectedPos pointing to the declared result type.
func TestReturn_ExpectedPos_Present(t *testing.T) {
    code := "package app\n" +
        "func F() (Owned<T>) { var x Owned<int,string>; return x }\n"
    var fs source.FileSet
    f := fs.AddFile("ret_expectedpos.ami", code)
    p := parser.New(f)
    af, _ := p.ParseFileCollect()
    ds := AnalyzeReturnTypes(af)
    found := false
    for _, d := range ds {
        if d.Code == "E_GENERIC_ARITY_MISMATCH" && d.Data != nil {
            if ep, ok := d.Data["expectedPos"].(diag.Position); ok {
                if ep.Line > 0 { found = true; break }
            }
        }
    }
    if !found {
        t.Fatalf("expected E_GENERIC_ARITY_MISMATCH with expectedPos; got %+v", ds)
    }
}

