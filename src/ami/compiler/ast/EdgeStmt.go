package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// EdgeStmt represents a simple edge: From -> To
type EdgeStmt struct {
	Pos     source.Position
	From    string
	FromPos source.Position
	To      string
	ToPos   source.Position
	Leading []Comment
}
