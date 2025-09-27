package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLintStageB_NoOpWhenEnabled(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "stageb_noop")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    setRuleToggles(RuleToggles{StageB: true, UnknownIdent: true, Unused: true, ImportExist: true, Duplicates: true, MemorySafety: true, RAIIHint: true})
    defer setRuleToggles(RuleToggles{})
    var buf bytes.Buffer
    if err := runLint(&buf, dir, false, false, false); err != nil {
        t.Fatalf("runLint: %v", err)
    }
    if !bytes.Contains(buf.Bytes(), []byte("lint: OK")) {
        t.Fatalf("expected OK with Stage B no-op; out=%s", buf.String())
    }
}

