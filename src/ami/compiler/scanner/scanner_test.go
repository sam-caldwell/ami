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
    if t1.Kind != token.Ident || t1.Lexeme != "package" { t.Fatalf("unexpected t1: %+v", t1) }
    t2 := s.Next()
    if t2.Kind != token.Ident || t2.Lexeme != "main" { t.Fatalf("unexpected t2: %+v", t2) }
    t3 := s.Next()
    if t3.Kind != token.EOF { t.Fatalf("expected EOF, got: %+v", t3) }
}

func TestScanner_NilOrEmpty(t *testing.T) {
    var s *Scanner
    if tok := s.Next(); tok.Kind != token.EOF { t.Fatalf("nil scanner should return EOF") }
    s = New(&source.File{Name: "t", Content: ""})
    if tok := s.Next(); tok.Kind != token.EOF { t.Fatalf("empty file should return EOF") }
}

