package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_UnmarkedMutatingAssignment_Errors(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "mut_assign")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "package x\nfunc F(){ var a; a = 1 }\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    setRuleToggles(RuleToggles{StageB: true, MemorySafety: true})
    defer setRuleToggles(RuleToggles{})
    var buf bytes.Buffer
    _ = runLint(&buf, dir, true, false, false)
    dec := json.NewDecoder(&buf)
    var saw bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        if m["code"] == "E_MUT_ASSIGN_UNMARKED" { saw = true }
    }
    if !saw { t.Fatalf("expected E_MUT_ASSIGN_UNMARKED; out=%s", buf.String()) }
}

