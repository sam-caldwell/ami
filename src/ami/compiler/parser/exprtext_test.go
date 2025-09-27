package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// TestExprText_CoversKinds exercises exprText for several expression node kinds.
func TestExprText_CoversKinds(t *testing.T) {
    // ident
    if exprText(&ast.IdentExpr{Name: "x"}) != "x" { t.Fatal("ident") }
    // string
    if exprText(&ast.StringLit{Value: "s"}) != "s" { t.Fatal("string") }
    // number
    if exprText(&ast.NumberLit{Text: "42"}) != "42" { t.Fatal("number") }
    // call
    if exprText(&ast.CallExpr{Name: "F"}) != "F()" { t.Fatal("call") }
    // selector chain
    x := &ast.SelectorExpr{Pos: source.Position{Offset: 1}, X: &ast.IdentExpr{Name: "a"}, Sel: "b"}
    if exprText(x) != "a.b" { t.Fatalf("selector: %q", exprText(x)) }
    // container literals map/set/slice
    if exprText(&ast.SliceLit{}) != "slice" { t.Fatal("slice") }
    if exprText(&ast.SetLit{}) != "set" { t.Fatal("set") }
    if exprText(&ast.MapLit{}) != "map" { t.Fatal("map") }
}

