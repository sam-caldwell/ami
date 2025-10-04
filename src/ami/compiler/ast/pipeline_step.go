package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// StepStmt represents a simple pipeline step call: Name(args...)
type StepStmt struct {
    Pos     source.Position
    Name    string
    Args    []Arg
    Leading []Comment
    Attrs   []Attr
    // Present when the step used call syntax. Zero when omitted.
    LParen  source.Position
    RParen  source.Position
}

func (*StepStmt) isNode() {}
func (*StepStmt) isStmt() {}
