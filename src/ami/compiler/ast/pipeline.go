package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// PipelineDecl represents a top-level pipeline declaration (scaffold).
type PipelineDecl struct {
    Pos  source.Position
    Name string
    Body *BlockStmt
    Error *ErrorBlock
    Leading []Comment
}

func (*PipelineDecl) isNode() {}
