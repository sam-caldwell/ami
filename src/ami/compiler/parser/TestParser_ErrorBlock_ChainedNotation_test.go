package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_ErrorBlock_ChainedNotation(t *testing.T) {
    src := "package app\npipeline P() { error { transform(a).collect().egress() } }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    pd, ok := file.Decls[0].(*ast.PipelineDecl)
    if !ok { t.Fatalf("decl type: %T", file.Decls[0]) }
    if pd.Error == nil || pd.Error.Body == nil { t.Fatalf("missing error block body") }
    var names []string
    for _, s := range pd.Error.Body.Stmts {
        if st, ok := s.(*ast.StepStmt); ok { names = append(names, st.Name) }
    }
    if len(names) != 3 { t.Fatalf("want 3 steps in error block, got %d (%v)", len(names), names) }
    want := []string{"transform", "collect", "egress"}
    for i := range want {
        if names[i] != want[i] { t.Fatalf("err step %d name: want %q, got %q", i, want[i], names[i]) }
    }
}

