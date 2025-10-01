package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Pipeline_Attr_KeyValueArgs(t *testing.T) {
    src := "package app\npipeline P() { Collect edge.MultiPath(window=100, backpressure=dropOldest); egress }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    pd, ok := file.Decls[0].(*ast.PipelineDecl)
    if !ok { t.Fatalf("decl0 not PipelineDecl: %T", file.Decls[0]) }
    var kvs []string
    for _, s := range pd.Stmts {
        if st, ok := s.(*ast.StepStmt); ok && st.Name == "Collect" {
            for _, at := range st.Attrs {
                if at.Name == "edge.MultiPath" && len(at.Args) >= 2 {
                    kvs = []string{at.Args[0].Text, at.Args[1].Text}
                }
            }
        }
    }
    if len(kvs) != 2 || kvs[0] != "window=100" || kvs[1] != "backpressure=dropOldest" {
        t.Fatalf("kv args: %+v", kvs)
    }
}

