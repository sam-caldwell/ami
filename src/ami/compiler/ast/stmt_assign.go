package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// AssignStmt represents a simple assignment: name = expr.
type AssignStmt struct {
    Pos     source.Position
    Name    string
    NamePos source.Position
    Value   Expr
    Leading []Comment
    Mutating bool            // true when assignment is marked with '*' on LHS
    StarPos  source.Position // position of '*', when Mutating
}

func (*AssignStmt) isNode() {}
func (*AssignStmt) isStmt() {}
