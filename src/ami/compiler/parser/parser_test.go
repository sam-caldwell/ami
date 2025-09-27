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
    // typed params/returns and statements: var, call, return
    src := "package app\nfunc F(a T, b U) (R1,R2) { var x T; Alpha(); return a,b }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
}

func TestParser_Pipeline_And_ErrorBlock(t *testing.T) {
    // Pipeline with inner error block; and a top-level error block
    src := "package app\n// leading\npipeline P() { error {} }\nerror {}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 2 { t.Fatalf("want 2 decls, got %d", len(file.Decls)) }
}

func TestParser_Tolerant_Recovery_MultipleErrors(t *testing.T) {
    // Two malformed lines: bad import path and bad func header; expect both errors collected
    src := "package app\nimport 123\nfunc ( {\n}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    _, errs := p.ParseFileCollect()
    if len(errs) < 2 {
        t.Fatalf("expected at least 2 errors, got %d: %+v", len(errs), errs)
    }
}

func TestParser_Pipeline_Steps_With_Args(t *testing.T) {
    src := "package app\npipeline P() {\n  // step 1\n  Alpha() attr1, attr2(\"p\")\n  Beta(\"x\", y) ;\n  A -> B;\n}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
}

func TestParser_Func_Assign_Binary_Defer(t *testing.T) {
    src := "package app\nfunc G(){ x = 1+2*3; defer Alpha(); }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    if _, err := p.ParseFile(); err != nil { t.Fatalf("parse: %v", err) }
}
