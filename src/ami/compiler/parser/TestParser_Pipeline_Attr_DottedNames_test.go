package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_Pipeline_Attr_DottedNames(t *testing.T) {
    src := "package app\npipeline P() { Alpha() edge.MultiPath(merge.Sort(\"ts\"), merge.Stable()) }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    // Verify attr args reduce to dotted names
    if len(file.Decls) != 1 { t.Fatalf("decls: %d", len(file.Decls)) }
    pd, ok := file.Decls[0].(*ast.PipelineDecl)
    if !ok { t.Fatalf("decl0: %T", file.Decls[0]) }
    var found bool
    for _, s := range pd.Stmts {
        if st, ok := s.(*ast.StepStmt); ok {
            for _, at := range st.Attrs {
                if at.Name == "edge.MultiPath" {
                    if len(at.Args) != 2 { t.Fatalf("args len: %d", len(at.Args)) }
                    if at.Args[0].Text != "merge.Sort(…)" || at.Args[1].Text != "merge.Stable()" {
                        t.Fatalf("attr args: %+v", at.Args)
                    }
                    found = true
                }
            }
        }
    }
    if !found { t.Fatalf("edge.MultiPath not found") }
}

