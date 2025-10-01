package scanner

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_tryScanTwoCharOperator(t *testing.T) {
    s := New(&source.File{Name: "t", Content: "=="})
    tok, ok := s.tryScanTwoCharOperator(s.file.Content, len(s.file.Content), 0)
    if !ok || tok.Kind != token.Eq || tok.Lexeme != "==" { t.Fatalf("twochar: %+v", tok) }
}

