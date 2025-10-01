package scanner

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_Next_DurationLiteralIntegration(t *testing.T) {
	s := New(&source.File{Name: "t", Content: "2h45m"})
	tok := s.Next()
	if tok.Kind != token.DurationLit || tok.Lexeme != "2h45m" {
		t.Fatalf("got %+v", tok)
	}
}
