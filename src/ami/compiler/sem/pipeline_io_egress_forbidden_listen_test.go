package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestPipelineSemantics_IO_EgressForbidden_Listen(t *testing.T) {
    f := &ast.File{PackageName: "app"}
    p := &ast.PipelineDecl{Name: "P"}
    p.Stmts = []ast.Stmt{
        &ast.StepStmt{Name: "ingress"},
        &ast.StepStmt{Name: "egress"},
        &ast.StepStmt{Name: "io.Listen"},
    }
    f.Decls = []ast.Decl{p}
    ds := AnalyzePipelineSemantics(f)
    found := false
    for _, d := range ds { if d.Code == "E_IO_PERMISSION" { found = true; break } }
    if !found { t.Fatalf("expected E_IO_PERMISSION for listen at egress; got %+v", ds) }
}
