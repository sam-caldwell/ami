package ast

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestEnumMember_struct(t *testing.T) {
	_ = EnumMember{
		Pos:  source.Position{},
		Name: string("test name"),
	}
}
