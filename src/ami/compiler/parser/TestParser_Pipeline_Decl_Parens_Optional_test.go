package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Ensure pipeline declarations work without an empty parameter list.
func TestParser_Pipeline_Decl_Parens_Optional(t *testing.T) {
    src := "package app\n" +
        "pipeline P { ingress; egress }\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
    pd, ok := file.Decls[0].(*ast.PipelineDecl)
    if !ok { t.Fatalf("decl not PipelineDecl: %T", file.Decls[0]) }
    if pd.Name != "P" { t.Fatalf("pipeline name: want P, got %s", pd.Name) }
    // Expect two step statements: ingress and egress
    var steps []string
    for _, s := range pd.Stmts {
        if st, ok := s.(*ast.StepStmt); ok { steps = append(steps, st.Name) }
    }
    if len(steps) != 2 || steps[0] != "ingress" || steps[1] != "egress" {
        t.Fatalf("unexpected steps: %v", steps)
    }
    // Paren positions should be zero (absent)
    if pd.LParen.Line != 0 || pd.RParen.Line != 0 {
        t.Fatalf("expected no paren positions, got L=%+v R=%+v", pd.LParen, pd.RParen)
    }
}

