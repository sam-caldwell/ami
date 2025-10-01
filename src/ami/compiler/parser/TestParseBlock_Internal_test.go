package parser

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// TestParseBlock_Internal exercises the private parseBlock depth logic.
func TestParseBlock_Internal(t *testing.T) {
	code := "{{}}"
	f := (&source.FileSet{}).AddFile("b.ami", code)
	p := New(f)
	if p.cur.Lexeme != "{" {
		t.Fatalf("want '{', got %q", p.cur.Lexeme)
	}
	if _, err := p.parseBlock(); err != nil {
		t.Fatalf("parseBlock: %v", err)
	}
}
