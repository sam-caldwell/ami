package scanner

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestScanner_FileContent(t *testing.T) {
	src := "package app\n"
	s := New(&source.File{Name: "t", Content: src})
	if s.FileContent() != src {
		t.Fatalf("FileContent mismatch")
	}
}
