package ast

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestAttr_Struct(t *testing.T) {
	_ = Attr{
		Pos: source.Position{
			Line:   int(1),
			Column: int(2),
			Offset: int(3),
		},
		Name: string("merge.Buffer"),
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
}
