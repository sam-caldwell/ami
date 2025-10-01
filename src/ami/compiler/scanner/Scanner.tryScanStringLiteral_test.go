package scanner

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_tryScanStringLiteral(t *testing.T) {
    s := New(&source.File{Name: "t", Content: `"hi"`})
    tok, ok := s.tryScanStringLiteral(s.file.Content, len(s.file.Content), 0)
    if !ok || tok.Kind != token.String || tok.Lexeme != `"hi"` { t.Fatalf("string: %+v", tok) }
    s = New(&source.File{Name: "t", Content: `"unterminated`})
    tok, ok = s.tryScanStringLiteral(s.file.Content, len(s.file.Content), 0)
    if !ok || tok.Kind != token.Unknown { t.Fatalf("unterminated: %+v", tok) }
}

