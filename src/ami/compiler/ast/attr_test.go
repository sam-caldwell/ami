package ast

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestAttr_Struct(t *testing.T) {
	a := Attr{
		Pos: source.Position{
			Line:   int(1),
			Column: int(2),
			Offset: int(3),
		},
		Name: "merge.Buffer",
		Args: []Arg{
			Arg{
				Pos: source.Position{
					Line:   int(4),
					Column: int(5),
					Offset: int(6),
				},
				Text:     string("sample text"),
				IsString: true,
			},
		},
	}
	if a.Name != "merge.Buffer" {
		t.Fatalf("unexpected: %+v", a)
	}
	if a.Pos.Line != int(1) {
		t.Errorf("Pos.Line should be 1")
	}
	if a.Pos.Column != int(2) {
		t.Errorf("Pos.Column should be 2")
	}
	if a.Pos.Offset != int(3) {
		t.Errorf("Pos.Offset should be 3")
	}
}
