package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_UnusedImport_Warns(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "unused_import")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "package x\nimport foo\nfunc F(){ }\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    setRuleToggles(RuleToggles{StageB: true, Unused: true})
    defer setRuleToggles(RuleToggles{})
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil {
        // unused import only triggers warn; ignore
    }
    dec := json.NewDecoder(&buf)
    var saw bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        if m["code"] == "W_UNUSED_IMPORT" { saw = true }
    }
    if !saw { t.Fatalf("expected W_UNUSED_IMPORT; out=%s", buf.String()) }
}

func TestLint_UnusedImport_Used_NoWarn(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "unused_import_used")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "package x\nimport foo\nfunc F(){ foo.Call() }\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    setRuleToggles(RuleToggles{StageB: true, Unused: true})
    defer setRuleToggles(RuleToggles{})
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil {
        // ignore
    }
    dec := json.NewDecoder(&buf)
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        if m["code"] == "W_UNUSED_IMPORT" { t.Fatalf("did not expect unused import warning: %s", buf.String()) }
    }
}

