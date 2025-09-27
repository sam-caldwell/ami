package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Validate that mod update reports selected versions for remote requirements that have entries in ami.sum.
func TestModUpdate_ReportsSelectedVersions_ForRemoteReqs(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_update", "select")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Workspace with a local package and a remote import constraint
    ws := workspace.DefaultWorkspace()
    ws.Packages = workspace.PackageList{
        {Key: "main", Package: workspace.Package{Name: "app", Version: "1.0.0", Root: "./src", Import: []string{"lib@^1.0.0"}}},
    }
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "x.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save ws: %v", err) }
    // Pre-existing ami.sum with multiple versions for lib
    sum := []byte(`{ "schema": "ami.sum/v1", "packages": { "lib": { "v1.2.3": "aaa", "v1.4.0": "bbb" } } }`)
    if err := os.WriteFile(filepath.Join(dir, "ami.sum"), sum, 0o644); err != nil { t.Fatalf("write sum: %v", err) }

    var buf bytes.Buffer
    if err := runModUpdate(&buf, dir, true); err != nil { t.Fatalf("runModUpdate: %v", err) }
    var res modUpdateResult
    if err := json.Unmarshal(buf.Bytes(), &res); err != nil { t.Fatalf("json: %v; out=%s", err, buf.String()) }
    // Expect selection of lib@v1.4.0 (highest satisfying ^1.0.0 without prerelease)
    found := false
    for _, s := range res.Selected { if s.Name == "lib" && s.Version == "v1.4.0" { found = true; break } }
    if !found { t.Fatalf("expected selection lib@v1.4.0; got: %+v", res.Selected) }
}

