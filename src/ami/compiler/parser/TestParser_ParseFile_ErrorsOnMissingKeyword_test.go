package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_ParseFile_ErrorsOnMissingKeyword(t *testing.T) {
    f := &source.File{Name: "t.ami", Content: "pkg app"}
    p := New(f)
    if _, err := p.ParseFile(); err == nil {
        t.Fatalf("expected error for missing 'package' keyword")
    }
}

