package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// VarDecl declares a local variable.
type VarDecl struct {
    Pos     source.Position
    Name    string
    NamePos source.Position
    Type    string
    TypePos source.Position
    Init    Expr // optional
    Leading []Comment
}

func (*VarDecl) isNode() {}
func (*VarDecl) isStmt() {}

