package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_FormattingMarkers_Warns(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "format")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // One line with trailing spaces and a tab-indented line
    content := "package x\n\tfunc F(){}   \n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil {
        // ignore exit code in JSON mode
    }
    dec := json.NewDecoder(&buf)
    var sawTab, sawTrail bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        if m["code"] == "W_FORMAT_TAB_INDENT" { sawTab = true }
        if m["code"] == "W_FORMAT_TRAILING_WS" { sawTrail = true }
    }
    if !sawTab || !sawTrail { t.Fatalf("expected formatting warnings; out=%s", buf.String()) }
}

