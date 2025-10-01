package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Pipeline_ChainedNotation(t *testing.T) {
    src := "package app\npipeline P() { ingress(a).transform(b,c).collect().egress() }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
    pd, ok := file.Decls[0].(*ast.PipelineDecl)
    if !ok { t.Fatalf("decl type: %T", file.Decls[0]) }
    // Expect four steps in order
    var names []string
    for _, s := range pd.Stmts {
        if st, ok := s.(*ast.StepStmt); ok { names = append(names, st.Name) }
    }
    if len(names) != 4 { t.Fatalf("want 4 steps, got %d (%v)", len(names), names) }
    want := []string{"ingress", "transform", "collect", "egress"}
    for i := range want {
        if names[i] != want[i] { t.Fatalf("step %d name: want %q, got %q", i, want[i], names[i]) }
    }
}

