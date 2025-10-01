package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Pipeline_EdgePipelineAttrParsed(t *testing.T) {
    src := "package app\npipeline P() { Alpha edge.Pipeline(name=X, type=\"Event<int>\"); egress }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    pd, ok := file.Decls[0].(*ast.PipelineDecl)
    if !ok { t.Fatalf("decl0 not PipelineDecl: %T", file.Decls[0]) }
    found := false
    for _, s := range pd.Stmts {
        if st, ok := s.(*ast.StepStmt); ok && st.Name == "Alpha" {
            for _, at := range st.Attrs { if at.Name == "edge.Pipeline" { found = true } }
        }
    }
    if !found { t.Fatalf("edge.Pipeline attr not found") }
}

