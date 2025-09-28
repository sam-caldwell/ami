package driver

import (
    "os"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestPipelinesDebug_MergeNorm_Determinism(t *testing.T) {
    pd := &ast.PipelineDecl{Name: "P"}
    st := &ast.StepStmt{Name: "Collect", Attrs: []ast.Attr{
        {Name: "merge.Sort", Args: []ast.Arg{{Text: "ts"}, {Text: "asc"}}},
        {Name: "merge.Buffer", Args: []ast.Arg{{Text: "8"}, {Text: "block"}}},
    }}
    pd.Stmts = []ast.Stmt{&ast.StepStmt{Name: "Ingress"}, st, &ast.StepStmt{Name: "Egress"}}
    f := &ast.File{Decls: []ast.Decl{pd}}
    p1, err := writePipelinesDebug("app", "u", f)
    if err != nil { t.Fatalf("write1: %v", err) }
    b1, _ := os.ReadFile(p1)
    p2, err := writePipelinesDebug("app", "u", f)
    if err != nil { t.Fatalf("write2: %v", err) }
    b2, _ := os.ReadFile(p2)
    if string(b1) != string(b2) {
        t.Fatalf("pipelines debug not deterministic for mergeNorm")
    }
}

