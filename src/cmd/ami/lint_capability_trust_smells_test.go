package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_Capability_Trust_Smells(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "caps_trust")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "package x\npipeline P() { ingress; io.Read(\"f\"); egress }\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    setRuleToggles(RuleToggles{StageB: true})
    defer setRuleToggles(RuleToggles{})
    var buf bytes.Buffer
    _ = runLint(&buf, dir, true, false, false)
    dec := json.NewDecoder(&buf)
    var sawCap, sawTrustUns, sawTrustWarn bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        switch m["code"] {
        case "W_CAPABILITY_UNDECLARED":
            sawCap = true
        case "W_TRUST_UNSPECIFIED":
            sawTrustUns = true
        case "W_TRUST_UNTRUSTED_IO":
            sawTrustWarn = true
        }
    }
    // trust warn only appears when level=untrusted; here we have unspecified
    if !(sawCap && sawTrustUns) { t.Fatalf("missing capability/trust smells; out=%s", buf.String()) }
    if sawTrustWarn { t.Fatalf("unexpected untrusted warning") }
}

