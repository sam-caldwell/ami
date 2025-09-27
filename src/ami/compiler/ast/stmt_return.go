package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// ReturnStmt represents a return statement with zero or more results.
type ReturnStmt struct {
    Pos     source.Position
    Results []Expr
    Leading []Comment
}

func (*ReturnStmt) isNode() {}
func (*ReturnStmt) isStmt() {}

