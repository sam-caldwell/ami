package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Pipeline_Steps_TypeAttrs_CollectEdges(t *testing.T) {
    src := "package app\npipeline P(){ ingress; A type(\"X\"); B type(\"Y\"); A -> Collect; B -> Collect; egress }\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    pd, ok := file.Decls[0].(*ast.PipelineDecl)
    if !ok { t.Fatalf("decl not PipelineDecl: %T", file.Decls[0]) }
    var aType, bType string
    for _, s := range pd.Stmts {
        if st, ok := s.(*ast.StepStmt); ok {
            if st.Name == "A" || st.Name == "B" {
                for _, at := range st.Attrs {
                    if at.Name == "type" && len(at.Args) > 0 { if st.Name == "A" { aType = at.Args[0].Text } else { bType = at.Args[0].Text } }
                }
            }
        }
    }
    if aType != "X" || bType != "Y" { t.Fatalf("types parsed: A=%q B=%q", aType, bType) }
}

