package driver

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLowerExpr_ConstFolding_Strings_Chained(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc F(){ (\"a\"+\"b\") + \"c\" }\n"
    fs.AddFile("folds.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    arts, _ := Compile(ws, pkgs, Options{Debug: true})
    if len(arts.IR) == 0 { t.Fatalf("no IR emitted") }
    b, err := os.ReadFile(arts.IR[0])
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fns := obj["functions"].([]any)
    fn := fns[0].(map[string]any)
    blks := fn["blocks"].([]any)
    blk := blks[0].(map[string]any)
    instrs := blk["instrs"].([]any)
    // Look for lit:"abc"
    found := false
    for _, in := range instrs {
        m := in.(map[string]any)
        if m["op"] == "EXPR" {
            e := m["expr"].(map[string]any)
            if op, _ := e["op"].(string); op == "lit:\"abc\"" { found = true; break }
        }
    }
    if !found { t.Fatalf("expected folded string literal lit:\"abc\"") }
}

