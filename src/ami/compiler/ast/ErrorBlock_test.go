package ast

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func Test_ErrorBlock_Exists(t *testing.T) {
	_ = ErrorBlock{
		Pos:  source.Position{},
		Body: &BlockStmt{},
		Leading: []Comment{
			Comment{},
			Comment{},
			Comment{},
		},
	}
}
