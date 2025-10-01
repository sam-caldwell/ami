package scanner

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestNew_ConstructsScanner(t *testing.T) {
	f := &source.File{Name: "t.ami", Content: "package x"}
	s := New(f)
	if s == nil {
		t.Fatalf("New returned nil")
	}
	if s.file != f {
		t.Fatalf("scanner.file mismatch")
	}
	if s.offset != 0 {
		t.Fatalf("scanner.offset want 0 got %d", s.offset)
	}
}
