package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure syntax errors are streamed as diag.v1 records in JSON mode.
func TestRunBuild_JSON_SyntaxDiagnostics_Streamed(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "json_syn_diags")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    // syntactically invalid: missing package keyword
    if err := os.WriteFile(filepath.Join(dir, "src", "bad.ami"), []byte("pkg app\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var out bytes.Buffer
    err := runBuild(&out, dir, true, false)
    if err == nil { t.Fatalf("expected error with syntax diagnostics") }
    s := bufio.NewScanner(bytes.NewReader(out.Bytes()))
    if !s.Scan() { t.Fatalf("expected at least one diag record") }
    var m map[string]any
    if e := json.Unmarshal(s.Bytes(), &m); e != nil { t.Fatalf("json: %v; line=%s", e, s.Text()) }
    if m["schema"] != "diag.v1" || m["code"] != "E_PARSE_SYNTAX" { t.Fatalf("diag: %+v", m) }
}

