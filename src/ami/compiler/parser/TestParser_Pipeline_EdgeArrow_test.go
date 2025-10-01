package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Pipeline_EdgeArrow(t *testing.T) {
    src := "package app\npipeline P(){ Alpha -> Beta; egress }\n"
    f := (&source.FileSet{}).AddFile("edge.ami", src)
    p := New(f)
    if _, err := p.ParseFile(); err != nil { t.Fatalf("ParseFile: %v", err) }
}

