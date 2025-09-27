package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_CLIParity_JSONVsHuman_SummaryCounts(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "parity")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte("UNKNOWN_IDENT\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{} // non-strict
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    // JSON mode: capture summary counts
    var jb bytes.Buffer
    if err := runLint(&jb, dir, true, false, false); err != nil { t.Fatalf("json lint err: %v", err) }
    dec := json.NewDecoder(&jb)
    var jErrs, jWarns int
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        if m["code"] == "SUMMARY" {
            if d, ok := m["data"].(map[string]any); ok {
                if v, ok := d["errors"].(float64); ok { jErrs = int(v) }
                if v, ok := d["warnings"].(float64); ok { jWarns = int(v) }
            }
            break
        }
    }
    if jErrs != 0 || jWarns < 1 { t.Fatalf("expected at least 1 warning in JSON; got e=%d,w=%d", jErrs, jWarns) }

    // Human mode: check summary line
    var hb bytes.Buffer
    if err := runLint(&hb, dir, false, false, false); err != nil { t.Fatalf("human lint err: %v", err) }
    s := hb.String()
    if !strings.Contains(s, "warning(s)") { t.Fatalf("expected summary in human out: %s", s) }
}

