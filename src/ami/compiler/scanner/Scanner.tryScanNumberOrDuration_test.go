package scanner

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_tryScanNumberOrDuration(t *testing.T) {
    // duration
    s := New(&source.File{Name: "t", Content: "2h45m"})
    tok, ok := s.tryScanNumberOrDuration(s.file.Content, len(s.file.Content), 0)
    if !ok || tok.Kind != token.DurationLit { t.Fatalf("duration: %+v", tok) }
    // number
    s = New(&source.File{Name: "t", Content: "123"})
    tok, ok = s.tryScanNumberOrDuration(s.file.Content, len(s.file.Content), 0)
    if !ok || tok.Kind != token.Number || tok.Lexeme != "123" { t.Fatalf("number: %+v", tok) }
}

