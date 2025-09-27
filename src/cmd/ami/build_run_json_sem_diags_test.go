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

// In JSON mode, compiler semantic diagnostics should be streamed and exit with USER error.
func TestRunBuild_JSON_SemanticDiagnostics_Streamed(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "json_sem_diags")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    // Syntactically valid but semantically invalid pipeline (no ingress at start)
    src := "package app\npipeline P(){ work; egress }\n"
    if err := os.WriteFile(filepath.Join(dir, "src", "bad.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    var out bytes.Buffer
    err := runBuild(&out, dir, true, false)
    if err == nil { t.Fatalf("expected error due to diagnostics") }
    // Expect at least one diag.v1 JSON record on stdout
    var first map[string]any
    s := bufio.NewScanner(bytes.NewReader(out.Bytes()))
    if !s.Scan() { t.Fatalf("expected at least one diag record; got none") }
    if e := json.Unmarshal(s.Bytes(), &first); e != nil { t.Fatalf("json: %v; line=%s", e, s.Text()) }
    if first["schema"] != "diag.v1" { t.Fatalf("schema: %v", first["schema"]) }
}

