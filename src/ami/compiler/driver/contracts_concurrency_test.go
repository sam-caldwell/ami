package driver

import (
    "encoding/json"
    "os"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// Ensure concurrency pragmas are reflected in contracts debug output.
func TestContractsDebug_IncludesConcurrencyPragmas(t *testing.T) {
    code := "package app\n#pragma concurrency:workers 4\n#pragma concurrency:schedule fair\n"
    f := &source.File{Name: "u.ami", Content: code}
    path, err := writeContractsDebug("app", "u", mustParse(t, f))
    if err != nil { t.Fatalf("contracts: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    conc, ok := obj["concurrency"].(map[string]any)
    if !ok { t.Fatalf("missing concurrency: %v", obj) }
    if int(conc["workers"].(float64)) != 4 || conc["schedule"].(string) != "fair" {
        t.Fatalf("concurrency mismatch: %+v", conc)
    }
}

func mustParse(t *testing.T, f *source.File) *ast.File {
    t.Helper()
    p := parser.New(f)
    af, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    return af
}
