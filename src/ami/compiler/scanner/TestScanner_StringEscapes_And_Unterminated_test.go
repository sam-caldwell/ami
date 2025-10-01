package scanner

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_StringEscapes_And_Unterminated(t *testing.T) {
	s := New(&source.File{Name: "t", Content: `"a\"b"`})
	tok := s.Next()
	if tok.Kind != token.String || tok.Lexeme != `"a\"b"` {
		t.Fatalf("escaped string: %+v", tok)
	}
	s = New(&source.File{Name: "t", Content: `"unterminated`})
	tok = s.Next()
	if tok.Kind != token.Unknown {
		t.Fatalf("unterminated should be Unknown: %+v", tok)
	}
}
