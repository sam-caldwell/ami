package parser

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

func (e SyntaxError) Position() source.Position { return e.Pos }

