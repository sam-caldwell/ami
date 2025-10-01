package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestPipelineSemantics_IO_IngressForbidden_Connect(t *testing.T) {
    f := &ast.File{PackageName: "app"}
    p := &ast.PipelineDecl{Name: "P"}
    p.Stmts = []ast.Stmt{
        &ast.StepStmt{Name: "io.Connect"},
        &ast.StepStmt{Name: "ingress"},
        &ast.StepStmt{Name: "egress"},
    }
    f.Decls = []ast.Decl{p}
    ds := AnalyzePipelineSemantics(f)
    found := false
    for _, d := range ds { if d.Code == "E_IO_PERMISSION" { found = true; break } }
    if !found { t.Fatalf("expected E_IO_PERMISSION for connect at ingress; got %+v", ds) }
}
