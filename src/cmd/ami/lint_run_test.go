package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunLint_OK_OnDefaultWorkspace(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "ok")
    if err := os.MkdirAll(dir, 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    ws := workspace.DefaultWorkspace()
    ws.Toolchain.Linter.Options = []string{} // ensure non-strict for this test
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
        t.Fatalf("save: %v", err)
    }
    var buf bytes.Buffer
    if err := runLint(&buf, dir, false, false, false); err != nil {
        t.Fatalf("runLint: %v", err)
    }
    if !bytes.Contains(buf.Bytes(), []byte("lint: OK")) {
        t.Fatalf("expected OK, got: %s", buf.String())
    }
}

func TestRunLint_JSON_MissingWorkspace(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "missing")
    if err := os.MkdirAll(dir, 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err == nil {
        t.Fatalf("expected error for missing workspace")
    }
    var m map[string]any
    if e := json.Unmarshal(buf.Bytes(), &m); e != nil {
        t.Fatalf("json decode: %v; out=%s", e, buf.String())
    }
    if m["code"] != "E_WS_MISSING" {
        t.Fatalf("expected code E_WS_MISSING; got %v", m["code"])
    }
}

func TestRunLint_Human_WithWarningsAndSummary(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "human_warn")
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Write UNKNOWN_IDENT to trigger a warning in human mode
    if err := os.WriteFile(filepath.Join(dir, "src", "main.ami"), []byte("UNKNOWN_IDENT\n"), 0o644); err != nil {
        t.Fatalf("write: %v", err)
    }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Name = "bad_name" // triggers naming warning
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{} // non-strict path
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
        t.Fatalf("save: %v", err)
    }
    var buf bytes.Buffer
    // Non-strict: expect human warnings and summary
    if err := runLint(&buf, dir, false, false, false); err != nil { t.Fatalf("runLint: %v", err) }
    out := buf.String()
    if !bytes.Contains([]byte(out), []byte("lint: warn W_UNKNOWN_IDENT")) { t.Fatalf("expected unknown ident warning, got: %s", out) }
    if !bytes.Contains([]byte(out), []byte("warning(s)")) { t.Fatalf("expected summary line, got: %s", out) }

    // Strict mode: expect non-nil error (human or JSON); choose human here
    buf.Reset()
    if err := runLint(&buf, dir, false, false, true); err == nil {
        t.Fatalf("expected error in strict mode when warnings exist")
    }
}
