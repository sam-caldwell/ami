package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Pipeline_Collect_InArg_MultiPath(t *testing.T) {
    src := "package app\npipeline P() { Collect(in=edge.MultiPath(merge.Sort(\"ts\"))); egress }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    pd, ok := file.Decls[0].(*ast.PipelineDecl)
    if !ok { t.Fatalf("decl0 not PipelineDecl: %T", file.Decls[0]) }
    var args []string
    for _, s := range pd.Stmts {
        if st, ok := s.(*ast.StepStmt); ok && st.Name == "Collect" {
            for _, a := range st.Args { args = append(args, a.Text) }
        }
    }
    if len(args) != 1 || args[0] != "in=edge.MultiPath(…)" {
        t.Fatalf("collect args: %+v", args)
    }
}

