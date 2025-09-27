package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_DuplicateFunctionDecl_Warns(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "dupfunc")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Two files declaring the same function name
    f1 := "package x\nfunc F(){}\n"
    f2 := "package x\nfunc F(){}\n"
    if err := os.WriteFile(filepath.Join(srcDir, "a.ami"), []byte(f1), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := os.WriteFile(filepath.Join(srcDir, "b.ami"), []byte(f2), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    setRuleToggles(RuleToggles{StageB: true})
    defer setRuleToggles(RuleToggles{})
    var buf bytes.Buffer
    _ = runLint(&buf, dir, true, false, false)
    dec := json.NewDecoder(&buf)
    var saw bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        if m["code"] == "W_DUP_FUNC_DECL" { saw = true }
    }
    if !saw { t.Fatalf("expected W_DUP_FUNC_DECL; out=%s", buf.String()) }
}

