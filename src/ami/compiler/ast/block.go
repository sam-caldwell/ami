package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// BlockStmt represents a block delimited by braces.
type BlockStmt struct {
    LBrace source.Position
    RBrace source.Position
}

func (*BlockStmt) isNode() {}

