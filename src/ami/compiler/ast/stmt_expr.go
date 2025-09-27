package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// ExprStmt wraps an expression as a statement.
type ExprStmt struct {
    Pos     source.Position
    X       Expr
    Leading []Comment
}

func (*ExprStmt) isNode() {}
func (*ExprStmt) isStmt() {}

