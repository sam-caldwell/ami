package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestEpos_MultipleKinds(t *testing.T) {
    pos := source.Position{Line: 1, Column: 1, Offset: 1}
    cases := []ast.Expr{
        &ast.IdentExpr{Pos: pos},
        &ast.StringLit{Pos: pos},
        &ast.NumberLit{Pos: pos},
        &ast.CallExpr{Pos: pos},
        &ast.BinaryExpr{Pos: pos},
        &ast.SliceLit{Pos: pos},
        &ast.SetLit{Pos: pos},
        &ast.MapLit{Pos: pos},
        &ast.SelectorExpr{Pos: pos},
    }
    for _, e := range cases {
        p := epos(e)
        if p.Offset != pos.Offset { t.Fatalf("unexpected pos: %+v", p) }
    }
}

