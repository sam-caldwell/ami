package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Pipeline_And_ErrorBlock(t *testing.T) {
    // Pipeline with inner error block capturing steps; and a top-level error block
    src := "package app\n// leading\npipeline P() { error { Alpha(); Beta() } }\nerror {}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 2 { t.Fatalf("want 2 decls, got %d", len(file.Decls)) }
    pd, ok := file.Decls[0].(*ast.PipelineDecl)
    if !ok { t.Fatalf("decl0 type: %T", file.Decls[0]) }
    if pd.Error == nil || pd.Error.Body == nil || len(pd.Error.Body.Stmts) == 0 { t.Fatalf("error block not captured with steps") }
}

