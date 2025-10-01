package parser

import (
    "os"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Fixture_EBNFSample(t *testing.T) {
    b, err := os.ReadFile("testdata/ebnf_sample.ami")
    if err != nil { t.Fatalf("read: %v", err) }
    f := &source.File{Name: "ebnf_sample.ami", Content: string(b)}
    p := New(f)
    if _, err := p.ParseFile(); err != nil { t.Fatalf("parse: %v", err) }
}

