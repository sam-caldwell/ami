package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_WorkspaceSchemaError_JSON(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "schema_json")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Toolchain.Compiler.Concurrency = "0" // invalid per schema
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var buf bytes.Buffer
    err := runBuild(&buf, dir, true)
    if err == nil { t.Fatalf("expected error") }
    // Decode diag record
    var m map[string]any
    if e := json.Unmarshal(buf.Bytes(), &m); e != nil {
        t.Fatalf("json: %v; out=%s", e, buf.String())
    }
    if m["schema"] != "diag.v1" { t.Fatalf("expected diag.v1; got %v", m["schema"]) }
    if m["code"] != "E_WS_SCHEMA" { t.Fatalf("expected E_WS_SCHEMA; got %v", m["code"]) }
    if m["file"] != "ami.workspace" { t.Fatalf("expected file ami.workspace; got %v", m["file"]) }
}

