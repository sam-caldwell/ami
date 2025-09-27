package parser

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_ParseFile_PackageDecl(t *testing.T) {
    f := &source.File{Name: "t.ami", Content: "package app"}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if file.PackageName != "app" { t.Fatalf("want package app, got %q", file.PackageName) }
}

func TestParser_ParseFile_ErrorsOnMissingKeyword(t *testing.T) {
    f := &source.File{Name: "t.ami", Content: "pkg app"}
    p := New(f)
    if _, err := p.ParseFile(); err == nil {
        t.Fatalf("expected error for missing 'package' keyword")
    }
}

