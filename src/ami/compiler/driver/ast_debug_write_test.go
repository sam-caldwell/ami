package driver

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestWriteASTDebug_WritesSummary(t *testing.T) {
    // Minimal AST file with one import and one simple function w/o body
    f := &ast.File{}
    f.Decls = append(f.Decls, &ast.ImportDecl{Path: "modA", Constraint: "^1.2.3"})
    f.Decls = append(f.Decls, &ast.FuncDecl{Name: "F", Params: []ast.Param{{Name: "a"}}, Results: []ast.Result{{Type: "int"}}})
    path, err := writeASTDebug("main", "u1", f)
    if err != nil { t.Fatalf("write: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read: %v", err) }
    var u struct{ Schema, Package, Unit string }
    if err := json.Unmarshal(b, &u); err != nil { t.Fatalf("json: %v", err) }
    if u.Schema != "ast.v1" || u.Package != "main" || u.Unit != "u1" { t.Fatalf("unexpected header: %+v", u) }
}

