package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestEPos_VariousKinds(t *testing.T) {
    pos := source.Position{Offset: 7}
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
    for _, e := range cases { if ePos(e).Offset != pos.Offset { t.Fatal("ePos") } }
}

