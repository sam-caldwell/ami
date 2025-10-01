package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

func TestPipelineSemantics_IO_InMiddle_Error(t *testing.T) {
    f := &ast.File{PackageName: "app"}
    p := &ast.PipelineDecl{Name: "P"}
    p.Stmts = []ast.Stmt{
        &ast.StepStmt{Name: "ingress"},
        &ast.StepStmt{Name: "io.Read"},
        &ast.StepStmt{Name: "egress"},
    }
    f.Decls = []ast.Decl{p}
    ds := AnalyzePipelineSemantics(f)
    var out []diag.Record
    out = append(out, ds...)
    found := false
    for _, d := range out { if d.Code == "E_IO_PERMISSION" { found = true } }
    if !found { t.Fatalf("expected E_IO_PERMISSION; got %+v", out) }
}
