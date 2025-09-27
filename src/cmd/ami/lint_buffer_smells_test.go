package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_Pipeline_BufferSmells(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "buffer_smell")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Use attribute forms Buffer(...) and merge.Buffer(...)
    content := "package x\npipeline P(){ merge(); merge.Buffer(1, dropNewest); A(); A Buffer(1, drop); }\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    setRuleToggles(RuleToggles{StageB: true})
    defer setRuleToggles(RuleToggles{})
    var buf bytes.Buffer
    _ = runLint(&buf, dir, true, false, false)
    dec := json.NewDecoder(&buf)
    var sawPolicy, sawAlias bool
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        switch m["code"] {
        case "W_PIPELINE_BUFFER_POLICY_SMELL":
            sawPolicy = true
        case "W_PIPELINE_BUFFER_DROP_ALIAS":
            sawAlias = true
        }
    }
    if !sawPolicy || !sawAlias { t.Fatalf("expected buffer smell diags; out=%s", buf.String()) }
}

