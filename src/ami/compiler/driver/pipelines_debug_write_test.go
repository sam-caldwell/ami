package driver

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestWritePipelinesDebug_WritesList(t *testing.T) {
    // Create a minimal pipeline AST with two steps and an edge
    pd := &ast.PipelineDecl{Name: "Pipe"}
    pd.Stmts = append(pd.Stmts, &ast.StepStmt{Name: "Ingress"})
    pd.Stmts = append(pd.Stmts, &ast.EdgeStmt{From: "Ingress", To: "Egress"})
    pd.Stmts = append(pd.Stmts, &ast.StepStmt{Name: "Egress"})
    f := &ast.File{Decls: []ast.Decl{pd}}
    path, err := writePipelinesDebug("main", "u1", f)
    if err != nil { t.Fatalf("write: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read: %v", err) }
    var pl struct{ Schema, Package, Unit string }
    if err := json.Unmarshal(b, &pl); err != nil { t.Fatalf("json: %v", err) }
    if pl.Schema != "pipelines.v1" || pl.Package != "main" || pl.Unit != "u1" { t.Fatalf("unexpected header: %+v", pl) }
}

