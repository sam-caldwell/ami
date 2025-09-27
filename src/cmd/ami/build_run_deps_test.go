package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_FailsWhenSumMissing_JSON(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "deps_sum_missing")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    // add a remote requirement so audit is meaningful
    p := ws.FindPackage("main")
    p.Import = []string{"modA 1.2.3"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    var buf bytes.Buffer
    err := runBuild(&buf, dir, true, false)
    if err == nil { t.Fatalf("expected error") }
    var m map[string]any
    if e := json.Unmarshal(buf.Bytes(), &m); e != nil { t.Fatalf("json: %v; out=%s", e, buf.String()) }
    if m["code"] != "E_INTEGRITY" { t.Fatalf("expected E_INTEGRITY; got %v", m["code"]) }
    data := m["data"].(map[string]any)
    if data == nil || data["sumFound"] != false { t.Fatalf("expected sumFound=false; data=%v", data) }
}
