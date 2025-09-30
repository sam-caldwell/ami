package driver

import (
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestContractsDebug_Deterministic(t *testing.T) {
    step1 := &ast.StepStmt{Name: "ingress"}
    step2 := &ast.StepStmt{Name: "Transform", Attrs: []ast.Attr{{Name: "type", Args: []ast.Arg{{Text: "int"}}}}}
    step3 := &ast.StepStmt{Name: "egress"}
    pd := &ast.PipelineDecl{Name: "P", Stmts: []ast.Stmt{step1, step2, step3}}
    f := &ast.File{Decls: []ast.Decl{pd}}
    p1, err := writeContractsDebug("app", "u", f)
    if err != nil { t.Fatalf("write1: %v", err) }
    b1, _ := os.ReadFile(p1)
    p2, err := writeContractsDebug("app", "u", f)
    if err != nil { t.Fatalf("write2: %v", err) }
    b2, _ := os.ReadFile(p2)
    if string(b1) != string(b2) { t.Fatalf("contracts not deterministic") }
}

