package ir

import (
    "testing"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestFromAST_ToSchema_SortedFunctions(t *testing.T) {
    f := &astpkg.File{Package: "p", Decls: []astpkg.Node{
        astpkg.FuncDecl{Name: "z"},
        astpkg.FuncDecl{Name: "a"},
        astpkg.FuncDecl{Name: "m"},
    }}
    m := FromASTFile("p", "unit.ami", f)
    ir := m.ToSchema()
    if len(ir.Functions) != 3 { t.Fatalf("functions=%d", len(ir.Functions)) }
    if ir.Functions[0].Name != "a" || ir.Functions[1].Name != "m" || ir.Functions[2].Name != "z" {
        t.Fatalf("not sorted: %v", []string{ir.Functions[0].Name, ir.Functions[1].Name, ir.Functions[2].Name})
    }
}

