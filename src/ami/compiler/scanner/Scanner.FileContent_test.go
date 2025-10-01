package scanner

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestScanner_FileContent_Method(t *testing.T) {
	// nil receiver
	var s *Scanner
	if s.FileContent() != "" {
		t.Fatalf("nil receiver should return empty content")
	}
	// nil file
	s = &Scanner{}
	if s.FileContent() != "" {
		t.Fatalf("nil file should return empty content")
	}
	// content passthrough
	s = New(&source.File{Name: "t", Content: "abc"})
	if got := s.FileContent(); got != "abc" {
		t.Fatalf("content mismatch: %q", got)
	}
}
