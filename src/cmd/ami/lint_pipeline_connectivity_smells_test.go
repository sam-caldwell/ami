package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_Pipeline_Connectivity_Smells(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "pipeline_conn")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // One unit with: ingress->A, and disconnected B; no path to egress
    content := "package x\npipeline P() { ingress; A; B; egress; ingress -> A; }\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    setRuleToggles(RuleToggles{StageB: true})
    defer setRuleToggles(RuleToggles{})
    var buf bytes.Buffer
    _ = runLint(&buf, dir, true, false, false)
    dec := json.NewDecoder(&buf)
    var sawNonterm, sawDisc, sawNoPath bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        switch m["code"] {
        case "W_PIPELINE_NONTERMINATING_NODE":
            sawNonterm = true
        case "W_PIPELINE_DISCONNECTED_NODE":
            sawDisc = true
        case "W_PIPELINE_NO_PATH_INGRESS_EGRESS":
            sawNoPath = true
        }
    }
    if !(sawNonterm && sawDisc && sawNoPath) {
        t.Fatalf("missing smells; nonterm=%v disc=%v nopath=%v out=%s", sawNonterm, sawDisc, sawNoPath, buf.String())
    }
}
