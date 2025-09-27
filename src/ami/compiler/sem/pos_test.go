package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestEpos_ReturnsExprPosition(t *testing.T) {
    // String literal
    s := &ast.StringLit{Pos: source.Position{Line: 2, Column: 3, Offset: 5}}
    if p := epos(s); p.Line != 2 || p.Column != 3 || p.Offset != 5 {
        t.Fatalf("unexpected pos: %+v", p)
    }
    // Binary expr uses its own Pos field
    be := &ast.BinaryExpr{Pos: source.Position{Line: 4, Column: 1, Offset: 7}}
    if p := epos(be); p.Line != 4 || p.Column != 1 || p.Offset != 7 {
        t.Fatalf("unexpected pos for binary: %+v", p)
    }
}

