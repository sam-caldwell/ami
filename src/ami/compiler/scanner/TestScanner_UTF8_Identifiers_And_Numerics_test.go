package scanner

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_UTF8_Identifiers_And_Numerics(t *testing.T) {
	src := "π = 3.14 0x1f 0b1010 0o77 1e9 2.5e-3"
	s := New(&source.File{Name: "t", Content: src})
	// π
	if tok := s.Next(); tok.Kind != token.Ident || tok.Lexeme != "π" {
		t.Fatalf("utf8 ident: %+v", tok)
	}
	// '='
	if tok := s.Next(); tok.Kind != token.Assign {
		t.Fatalf("assign: %+v", tok)
	}
	// decimals and variants
	want := []string{"3.14", "0x1f", "0b1010", "0o77", "1e9", "2.5e-3"}
	for _, w := range want {
		tok := s.Next()
		if tok.Kind != token.Number || tok.Lexeme != w {
			t.Fatalf("num %q => %+v", w, tok)
		}
	}
}
