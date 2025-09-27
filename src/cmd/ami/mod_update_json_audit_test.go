package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestModUpdate_EmbedsAuditInJSON(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_update", "audit_json")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages = workspace.PackageList{
        {Key: "main", Package: workspace.Package{Name: "app", Version: "1.0.0", Root: "./src", Import: []string{"lib@^1.2.3"}}},
    }
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "a.txt"), []byte("a"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    // No ami.sum present -> SumFound false; MissingInSum should include lib
    var buf bytes.Buffer
    if err := runModUpdate(&buf, dir, true); err != nil { t.Fatalf("runModUpdate: %v", err) }
    var res modUpdateResult
    if err := json.Unmarshal(buf.Bytes(), &res); err != nil { t.Fatalf("json: %v; out=%s", err, buf.String()) }
    if res.Audit == nil { t.Fatalf("expected audit object present") }
    if res.Audit.SumFound { t.Fatalf("expected SumFound=false") }
}

