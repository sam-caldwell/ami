package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// ErrorBlock represents a top-level error block (scaffold).
type ErrorBlock struct {
	Pos     source.Position
	Body    *BlockStmt
	Leading []Comment
}
