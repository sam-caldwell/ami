package scanner

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_NilOrEmpty(t *testing.T) {
	var s *Scanner
	if tok := s.Next(); tok.Kind != token.EOF {
		t.Fatalf("nil scanner should return EOF")
	}
	s = New(&source.File{Name: "t", Content: ""})
	if tok := s.Next(); tok.Kind != token.EOF {
		t.Fatalf("empty file should return EOF")
	}
}
