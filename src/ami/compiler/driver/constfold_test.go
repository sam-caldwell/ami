package driver

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLowerExpr_ConstFolding_Binary(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // Expression statement and a return with a folded literal
    code := "package app\nfunc F() (int) { 1+2; return 1+2 }\n"
    fs.AddFile("fold.ami", code)
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
    // First EXPR should be a literal op lit:3
    foundExpr := false
    for _, ins := range instrs {
        m := ins.(map[string]any)
        if m["op"] == "EXPR" {
            expr := m["expr"].(map[string]any)
            if op, _ := expr["op"].(string); len(op) >= 4 && op[:4] == "lit:" {
                foundExpr = true
                break
            }
        }
    }
    if !foundExpr { t.Fatalf("expected folded literal EXPR in IR") }
    // Last instruction RETURN should carry int type
    last := instrs[len(instrs)-1].(map[string]any)
    if last["op"] != "RETURN" { t.Fatalf("last op: %v", last["op"]) }
    vals := last["values"].([]any)
    v0 := vals[0].(map[string]any)
    if v0["type"] != "int" { t.Fatalf("return type: %v", v0["type"]) }
}
