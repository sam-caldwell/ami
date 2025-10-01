package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Check IR JSON encodes select expression with expected types
func TestLower_Ternary_IR_JSON_Shapes_Select(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\nfunc F() (int){ var a int; var b int; return (a == 1) ? b : 2 }\n"
    fs.AddFile("u.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    path := filepath.Join("build", "debug", "ir", "app", "u.ir.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fns := obj["functions"].([]any)
    found := false
    for _, fnv := range fns {
        fn := fnv.(map[string]any)
        blks := fn["blocks"].([]any)
        for _, bv := range blks {
            blk := bv.(map[string]any)
            ins := blk["instrs"].([]any)
            for _, iv := range ins {
                m := iv.(map[string]any)
                if m["op"] == "COND_BR" || m["op"] == "PHI" {
                    found = true
                }
            }
        }
    }
    if !found { t.Fatalf("missing select in IR JSON") }
}
