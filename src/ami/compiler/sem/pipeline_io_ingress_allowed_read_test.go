package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestPipelineSemantics_IO_IngressAllowed_Read(t *testing.T) {
    f := &ast.File{PackageName: "app"}
    p := &ast.PipelineDecl{Name: "P"}
    p.Stmts = []ast.Stmt{
        &ast.StepStmt{Name: "io.Read"},
        &ast.StepStmt{Name: "ingress"},
        &ast.StepStmt{Name: "egress"},
    }
    f.Decls = []ast.Decl{p}
    ds := AnalyzePipelineSemantics(f)
    for _, d := range ds { if d.Code == "E_IO_PERMISSION" { t.Fatalf("unexpected E_IO_PERMISSION: %+v", ds) } }
}
