package ast

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func Test_enum_Exists(t *testing.T) {
	_ = EnumDecl{
		Pos:     source.Position{},
		NamePos: source.Position{},
		Name:    string("test name"),
		LBrace:  source.Position{},
		Members: []EnumMember{
			EnumMember{},
			EnumMember{},
			EnumMember{},
		},
		RBrace: source.Position{},
	}
}
