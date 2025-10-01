package scanner

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_Numbers_Strings_Operators(t *testing.T) {
	src := `123 "hi" == != <= >= && || -> + - * / % !`
	s := New(&source.File{Name: "t", Content: src})
	// number
	if tok := s.Next(); tok.Kind != token.Number || tok.Lexeme != "123" {
		t.Fatalf("num: %+v", tok)
	}
	// string
	if tok := s.Next(); tok.Kind != token.String || tok.Lexeme != `"hi"` {
		t.Fatalf("str: %+v", tok)
	}
	// operators (2-char first)
	wantOps := []struct {
		lex string
		k   token.Kind
	}{
		{"==", token.Eq}, {"!=", token.Ne}, {"<=", token.Le}, {">=", token.Ge},
		{"&&", token.And}, {"||", token.Or}, {"->", token.Arrow},
		{"+", token.Plus}, {"-", token.Minus}, {"*", token.Star}, {"/", token.Slash}, {"%", token.Percent}, {"!", token.Bang},
	}
	for _, w := range wantOps {
		tok := s.Next()
		if tok.Kind != w.k || tok.Lexeme != w.lex {
			t.Fatalf("op %q => %+v", w.lex, tok)
		}
	}
	if tok := s.Next(); tok.Kind != token.EOF {
		t.Fatalf("expected EOF, got %+v", tok)
	}
}
