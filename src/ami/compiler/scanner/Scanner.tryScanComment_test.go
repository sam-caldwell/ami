package scanner

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func TestScanner_tryScanComment(t *testing.T) {
	// line comment
	s := New(&source.File{Name: "t", Content: "// hi\n"})
	tok, ok := s.tryScanComment(s.file.Content, len(s.file.Content), 0)
	if !ok || tok.Kind != token.LineComment || tok.Lexeme == "" {
		t.Fatalf("line comment scan failed: %+v", tok)
	}
	// block comment
	s = New(&source.File{Name: "t", Content: "/* block */"})
	tok, ok = s.tryScanComment(s.file.Content, len(s.file.Content), 0)
	if !ok || tok.Kind != token.BlockComment || tok.Lexeme == "" {
		t.Fatalf("block comment scan failed: %+v", tok)
	}
}
