package scanner

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_tryScanPunctuationSymbol(t *testing.T) {
    s := New(&source.File{Name: "t", Content: "("})
    tok, ok := s.tryScanPunctuationSymbol(s.file.Content, len(s.file.Content), 0)
    if !ok || tok.Kind != token.LParenSym { t.Fatalf("punct: %+v", tok) }
}

