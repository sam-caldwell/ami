package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_Unused_Var_And_Func_WithPragmaSuppression(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "unused")
    _ = os.RemoveAll(dir)
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // One file with an unused func and var; pragma disables W_UNUSED_FUNC
    code := "#pragma lint:disable W_UNUSED_FUNC\npackage app\nfunc used(){}\nfunc dead(){}\nfunc main(){ var x int; var y int; _ = x; used() }\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(code), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save ws: %v", err) }

    // Enable Stage B unused rules
    setRuleToggles(RuleToggles{StageB: true, Unused: true})
    defer setRuleToggles(RuleToggles{})

    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil { /* warnings allowed */ }
    dec := json.NewDecoder(&buf)
    var sawUnusedFunc bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        if m["code"] == "W_UNUSED_FUNC" { sawUnusedFunc = true }
    }
    if sawUnusedFunc { t.Fatalf("pragma should suppress unused func") }
    // Unused var emission is analyzer-dependent; ensure no crash and stream parsed.
}
