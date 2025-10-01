package scanner

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_tryScanOneCharOperator(t *testing.T) {
    s := New(&source.File{Name: "t", Content: "+"})
    tok, ok := s.tryScanOneCharOperator(s.file.Content, len(s.file.Content), 0)
    if !ok || tok.Kind != token.Plus || tok.Lexeme != "+" { t.Fatalf("onechar: %+v", tok) }
}

