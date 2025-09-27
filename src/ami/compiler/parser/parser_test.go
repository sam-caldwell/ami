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

func TestParser_Imports_And_Func(t *testing.T) {
    src := "package app\nimport alpha\nimport \"beta\"\nfunc main() {}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if file.PackageName != "app" { t.Fatalf("pkg: %q", file.PackageName) }
    if len(file.Decls) != 3 { t.Fatalf("want 3 decls (2 imports, 1 func), got %d", len(file.Decls)) }
}

func TestParser_Func_Params_Results_And_Body(t *testing.T) {
    src := "package app\nfunc F(a,b) (R1,R2) { x y z }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
}

func TestParser_Pipeline_And_ErrorBlock(t *testing.T) {
    src := "package app\npipeline P() {}\nerror {}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 2 { t.Fatalf("want 2 decls, got %d", len(file.Decls)) }
}
