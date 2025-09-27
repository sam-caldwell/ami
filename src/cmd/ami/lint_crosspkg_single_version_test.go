package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_CrossPackage_SingleVersionRule_StrictRequiresPin(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "crosspkg_single")
    if err := os.MkdirAll(filepath.Join(dir, "lib1"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.MkdirAll(filepath.Join(dir, "lib2"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    // Overlapping ranges but no exact pin
    ws.Packages = append(ws.Packages, workspace.PackageEntry{Key: "lib1", Package: workspace.Package{Name: "Lib1", Version: "0.1.0", Root: "./lib1", Import: []string{"github.com/acme/a ^1.2.0"}}})
    ws.Packages = append(ws.Packages, workspace.PackageEntry{Key: "lib2", Package: workspace.Package{Name: "Lib2", Version: "0.1.0", Root: "./lib2", Import: []string{"github.com/acme/a ~1.2.3"}}})
    ws.Toolchain.Linter.Options = []string{"strict"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var buf bytes.Buffer
    err := runLint(&buf, dir, true, false, false)
    if err == nil { t.Fatalf("expected error due to strict single-version rule") }
    dec := json.NewDecoder(&buf)
    var saw bool
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        if m["code"] == "W_IMPORT_SINGLE_VERSION" { saw = true }
    }
    if !saw { t.Fatalf("expected W_IMPORT_SINGLE_VERSION; out=%s", buf.String()) }
}

