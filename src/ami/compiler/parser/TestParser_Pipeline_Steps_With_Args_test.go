package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Pipeline_Steps_With_Args(t *testing.T) {
    src := "package app\npipeline P() {\n  // step 1\n  Alpha() attr1, attr2(\"p\")\n  Beta(\"x\", y) ;\n  A -> B;\n}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
}

