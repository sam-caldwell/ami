package ast

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func Test_comment_Exists(t *testing.T) {
	_ = Comment{
		Pos:  source.Position{},
		Text: string("test string"),
	}
}
