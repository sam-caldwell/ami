package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// IfStmt represents: if <cond> { ... } [else { ... }]
type IfStmt struct {
    Pos     source.Position
    Leading []Comment
    Cond    Expr
    Then    *BlockStmt
    Else    *BlockStmt // optional
}

func (*IfStmt) isNode() {}
func (*IfStmt) isStmt() {}

