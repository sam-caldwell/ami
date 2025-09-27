package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// SelectorExpr represents a.b form.
type SelectorExpr struct {
    Pos    source.Position // position of left expression
    X      Expr            // left side
    Sel    string          // selected identifier
    SelPos source.Position // position of selected ident
}

func (*SelectorExpr) isNode() {}
func (*SelectorExpr) isExpr() {}

