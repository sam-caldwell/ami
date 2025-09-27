package driver

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLowerExpr_ConstFolding_NumericBases_And_Precedence(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // chained/precedence: 1+2*3+0x10 => 23; and 0xA+0b1+0o7 => 18
    code := "package app\nfunc F(){ 1+2*3+0x10; 0xA+0b1+0o7 }\n"
    fs.AddFile("fold2.ami", code)
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
    // count folded literal exprs
    lits := 0
    for _, in := range instrs {
        m := in.(map[string]any)
        if m["op"] == "EXPR" {
            e := m["expr"].(map[string]any)
            if op, _ := e["op"].(string); len(op) >= 4 && op[:4] == "lit:" { lits++ }
        }
    }
    if lits < 2 { t.Fatalf("expected at least two folded literal exprs, got %d", lits) }
}

