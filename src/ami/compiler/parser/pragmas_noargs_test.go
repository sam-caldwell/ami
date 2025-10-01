package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Pragmas_NoArgs_HeadOnly(t *testing.T) {
    src := "package app\n#pragma simple\n"
    f := &source.File{Name: "p.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Pragmas) != 1 { t.Fatalf("want 1 pragma, got %d", len(file.Pragmas)) }
    if file.Pragmas[0].Domain != "simple" || file.Pragmas[0].Key != "" { t.Fatalf("pragma: %+v", file.Pragmas[0]) }
}

