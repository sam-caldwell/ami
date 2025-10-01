package scanner

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_CommentsAndSymbols(t *testing.T) {
	src := `// line comment
(/* block */) ,;|#`
	s := New(&source.File{Name: "t.ami", Content: src})
	// Expect line comment first
	t1 := s.Next()
	if t1.Kind != token.LineComment || t1.Lexeme == "" {
		t.Fatalf("want line comment, got %+v", t1)
	}
	// then LParen symbol
	t1 = s.Next()
	if t1.Kind != token.LParenSym || t1.Lexeme != token.LParen {
		t.Fatalf("want LParen sym, got %+v", t1)
	}
	// block comment next (inside parens)
	tbc := s.Next()
	if tbc.Kind != token.BlockComment || tbc.Lexeme == "" {
		t.Fatalf("want block comment, got %+v", tbc)
	}
	// RParen
	t2 := s.Next()
	if t2.Kind != token.RParenSym || t2.Lexeme != token.RParen {
		t.Fatalf("want RParen sym, got %+v", t2)
	}
	// comma
	t3 := s.Next()
	if t3.Kind != token.CommaSym || t3.Lexeme != token.Comma {
		t.Fatalf("want comma sym, got %+v", t3)
	}
	// semi
	t4 := s.Next()
	if t4.Kind != token.SemiSym || t4.Lexeme != token.Semi {
		t.Fatalf("want semicolon sym, got %+v", t4)
	}
	// '|' is now a bitwise operator
	if tok := s.Next(); tok.Kind != token.BitOr || tok.Lexeme != "|" {
		t.Fatalf("want bitwise '|' operator, got %+v", tok)
	}
	// pound
	if tok := s.Next(); tok.Kind != token.PoundSym || tok.Lexeme != "#" {
		t.Fatalf("want pound sym, got %+v", tok)
	}
	// EOF
	if tok := s.Next(); tok.Kind != token.EOF {
		t.Fatalf("expected EOF, got %+v", tok)
	}
}
