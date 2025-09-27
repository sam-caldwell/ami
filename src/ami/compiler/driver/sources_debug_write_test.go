package driver

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestWriteSourcesDebug_ProducesJSON(t *testing.T) {
    f := &ast.File{}
    f.Decls = append(f.Decls, &ast.ImportDecl{Path: "modA", Constraint: ">= 1.2.3"})
    f.Pragmas = append(f.Pragmas, ast.Pragma{Pos: ast.Position{Line: 1}, Domain: "edge", Key: "policy", Value: "dropNewest", Params: map[string]string{"cap":"1"}})
    path, err := writeSourcesDebug("main", "u1", f)
    if err != nil { t.Fatalf("write: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read: %v", err) }
    var obj struct{ Schema, Package, Unit string; ImportsDetailed []struct{ Path, Constraint string } }
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if obj.Schema != "sources.v1" || obj.Package != "main" || obj.Unit != "u1" { t.Fatalf("unexpected header: %+v", obj) }
    if len(obj.ImportsDetailed) != 1 || obj.ImportsDetailed[0].Path != "modA" { t.Fatalf("unexpected imports: %+v", obj.ImportsDetailed) }
}

