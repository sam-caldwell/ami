package scanner

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_scanFallbackSymbol(t *testing.T) {
    s := New(&source.File{Name: "t", Content: "§"})
    tok := s.scanFallbackSymbol(s.file.Content, len(s.file.Content), 0)
    if tok.Kind != token.Symbol || tok.Lexeme == "" { t.Fatalf("fallback: %+v", tok) }
}

