package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// DeferStmt represents a defer of a call expression.
type DeferStmt struct {
    Pos  source.Position
    Call *CallExpr
    Leading []Comment
}

func (*DeferStmt) isNode() {}
func (*DeferStmt) isStmt() {}

