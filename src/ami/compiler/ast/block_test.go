package ast

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func Test_block_Exists(t *testing.T) {
	_ = BlockStmt{
		LBrace: source.Position{
			Line:   int(4),
			Column: int(5),
			Offset: int(6),
		},
		RBrace: source.Position{
			Line:   int(7),
			Column: int(8),
			Offset: int(9),
		},
		Stmts: []Stmt{
			&AssignStmt{},
			&AssignStmt{},
			&AssignStmt{},
			&AssignStmt{},
		},
	}
}
