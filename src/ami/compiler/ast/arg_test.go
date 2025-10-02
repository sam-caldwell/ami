package ast

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func Test_arg_Exists(t *testing.T) {
	arg := Arg{
		Pos: source.Position{
			Line:   int(1),
			Column: int(2),
			Offset: int(3),
		},
		Text:     string("sample text"),
		IsString: true,
	}
	if !arg.IsString {
		t.Errorf("IsString should be true")
	}
	if arg.Text != "sample text" {
		t.Errorf("Text should be sample text")
	}
	if arg.Pos.Line != int(1) {
		t.Errorf("Pos.Line should be 1")
	}
	if arg.Pos.Column != int(2) {
		t.Errorf("Pos.Column should be 2")
	}
	if arg.Pos.Offset != int(3) {
		t.Errorf("Pos.Offset should be 3")
	}
}
