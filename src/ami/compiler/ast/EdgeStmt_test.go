package ast

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func Test_EdgeStmt_Exists(t *testing.T) {
	_ = EdgeStmt{
		Pos:     source.Position{},
		From:    string("test from string"),
		FromPos: source.Position{},
		To:      string("test to string"),
		ToPos:   source.Position{},
		Leading: []Comment{},
	}
}
