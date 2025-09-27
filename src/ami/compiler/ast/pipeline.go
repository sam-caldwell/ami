package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// PipelineDecl represents a top-level pipeline declaration (scaffold).
type PipelineDecl struct {
    Pos  source.Position
    Name string
    NamePos source.Position
    Body *BlockStmt
    Error *ErrorBlock
    Leading []Comment
    Stmts  []Stmt
    LParen source.Position
    RParen source.Position
}

func (*PipelineDecl) isNode() {}
