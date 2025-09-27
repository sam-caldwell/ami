package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// AssignStmt represents a simple assignment: name = expr.
type AssignStmt struct {
    Pos     source.Position
    Name    string
    NamePos source.Position
    Value   Expr
    Leading []Comment
}

func (*AssignStmt) isNode() {}
func (*AssignStmt) isStmt() {}

