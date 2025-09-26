package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "testing"
)

func TestSem_DuplicateTypeParams_Error(t *testing.T) {
    f := &astpkg.File{Package: "p", Decls: []astpkg.Node{
        astpkg.FuncDecl{Name: "f", TypeParams: []astpkg.TypeParam{{Name: "T"}, {Name: "T"}}},
    }}
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics {
        if d.Code == "E_DUP_TYPE_PARAM" {
            found = true
            break
        }
    }
    if !found {
        t.Fatalf("expected E_DUP_TYPE_PARAM, got: %+v", res.Diagnostics)
    }
}
