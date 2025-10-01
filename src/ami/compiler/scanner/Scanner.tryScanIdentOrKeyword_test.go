package scanner

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_tryScanIdentOrKeyword(t *testing.T) {
    s := New(&source.File{Name: "t", Content: "package"})
    tok, ok := s.tryScanIdentOrKeyword(s.file.Content, len(s.file.Content), 0)
    if !ok || tok.Kind != token.KwPackage { t.Fatalf("kw package: %+v", tok) }
    s = New(&source.File{Name: "t", Content: "main"})
    tok, ok = s.tryScanIdentOrKeyword(s.file.Content, len(s.file.Content), 0)
    if !ok || tok.Kind != token.Ident || tok.Lexeme != "main" { t.Fatalf("ident: %+v", tok) }
}

