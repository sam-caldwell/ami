package scanner

import (
	"testing"

	tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_TokensForPackageVersion(t *testing.T) {
	s := New("package main:0.1.2\n")
	t1 := s.Next()
	t2 := s.Next()
	t3 := s.Next()
	t4 := s.Next()
	t5 := s.Next()
	t6 := s.Next()
	if t1.Kind != tok.KW_PACKAGE {
		t.Fatalf("t1=%v", t1)
	}
	if t2.Kind != tok.IDENT || t2.Lexeme != "main" {
		t.Fatalf("t2=%v", t2)
	}
	if t3.Kind != tok.COLON {
		t.Fatalf("t3=%v", t3)
	}
	if t4.Kind != tok.NUMBER || t4.Lexeme != "0.1" {
		t.Fatalf("t4=%v", t4)
	}
	if t5.Kind != tok.DOT {
		t.Fatalf("t5=%v", t5)
	}
	if t6.Kind != tok.NUMBER || t6.Lexeme != "2" {
		t.Fatalf("t6=%v", t6)
	}
}
