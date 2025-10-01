package scanner

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_IdentAndEOF(t *testing.T) {
	f := &source.File{Name: "t.ami", Content: "package main"}
	s := New(f)
	t1 := s.Next()
	if t1.Kind != token.KwPackage || t1.Lexeme != "package" {
		t.Fatalf("unexpected t1: %+v", t1)
	}
	t2 := s.Next()
	if t2.Kind != token.Ident || t2.Lexeme != "main" {
		t.Fatalf("unexpected t2: %+v", t2)
	}
	t3 := s.Next()
	if t3.Kind != token.EOF {
		t.Fatalf("expected EOF, got: %+v", t3)
	}
}
