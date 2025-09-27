package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure mod update writes ami.sum in canonical nested format via workspace.Manifest
func TestModUpdate_WritesCanonicalNestedSum(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_update", "sum_format")
    src := filepath.Join(dir, "src")
    if err := os.MkdirAll(src, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(src, "x.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    ws := workspace.DefaultWorkspace()
    ws.Packages = workspace.PackageList{
        {Key: "main", Package: workspace.Package{Name: "app", Version: "0.0.1", Root: "./src"}},
    }
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    cache := filepath.Join(dir, "cache")
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", cache)

    var buf bytes.Buffer
    if err := runModUpdate(&buf, dir, true); err != nil { t.Fatalf("runModUpdate: %v", err) }
    // Read ami.sum and assert nested packages shape
    b, err := os.ReadFile(filepath.Join(dir, "ami.sum"))
    if err != nil { t.Fatalf("read sum: %v", err) }
    var m map[string]any
    if err := json.Unmarshal(b, &m); err != nil { t.Fatalf("json: %v", err) }
    pkgs, ok := m["packages"].(map[string]any)
    if !ok { t.Fatalf("packages not object form: %T", m["packages"]) }
    app, ok := pkgs["app"].(map[string]any)
    if !ok { t.Fatalf("app entry not map: %T", pkgs["app"]) }
    if _, ok := app["0.0.1"]; !ok { t.Fatalf("expected version key nested under app") }
}
