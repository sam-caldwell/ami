package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// E-8-2: Verify lowering expands a single multi-result call at a return site
// into multiple return values in IR JSON.
func TestReturnTuple_Expand_MultiResultCall_InIR(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nimport math\nfunc F() (float64, float64) { return math.Sincos(1.0) }\n"
    fs.AddFile("rtuple.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    // Semantics may emit return arity diags since stdlib signatures are not loaded as AMI sources.
    // We ignore diags here and focus on IR shape.
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "ir", "app", "rtuple.ir.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    // Find function F and its return instruction
    var foundReturn bool
    fns := obj["functions"].([]any)
    for _, fnv := range fns {
        fn := fnv.(map[string]any)
        if fn["name"].(string) != "F" { continue }
        blks := fn["blocks"].([]any)
        for _, bv := range blks {
            blk := bv.(map[string]any)
            for _, iv := range blk["instrs"].([]any) {
                in := iv.(map[string]any)
                if in["op"].(string) == "RETURN" {
                    vals := in["values"].([]any)
                    if len(vals) != 2 {
                        t.Fatalf("expected 2 return values, got %d: %+v", len(vals), vals)
                    }
                    // Types are present in value objects
                    v0 := vals[0].(map[string]any)["type"].(string)
                    v1 := vals[1].(map[string]any)["type"].(string)
                    if v0 != "float64" || v1 != "float64" {
                        t.Fatalf("return value types mismatch: %s,%s", v0, v1)
                    }
                    foundReturn = true
                }
            }
        }
    }
    if !foundReturn { t.Fatalf("RETURN not found in IR") }
}
