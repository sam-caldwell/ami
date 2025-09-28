package driver

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestContractsDebug_Write_Minimal(t *testing.T) {
    // Build a simple pipeline AST with a Transform(type=int)
    step1 := &ast.StepStmt{Name: "Ingress"}
    step2 := &ast.StepStmt{Name: "Transform", Attrs: []ast.Attr{{Name: "type", Args: []ast.Arg{{Text: "int"}}}}}
    step3 := &ast.StepStmt{Name: "Egress"}
    pd := &ast.PipelineDecl{Name: "P", Stmts: []ast.Stmt{step1, step2, step3}}
    f := &ast.File{Decls: []ast.Decl{pd}}
    path, err := writeContractsDebug("app", "u", f)
    if err != nil { t.Fatalf("write contracts: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if obj["schema"] != "contracts.v1" || obj["package"] != "app" { t.Fatalf("header: %#v", obj) }
}
