package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Imports_And_Func(t *testing.T) {
    src := "package app\nimport alpha\nimport \"beta\"\nfunc main() {}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if file.PackageName != "app" { t.Fatalf("pkg: %q", file.PackageName) }
    if len(file.Decls) != 3 { t.Fatalf("want 3 decls (2 imports, 1 func), got %d", len(file.Decls)) }
}

