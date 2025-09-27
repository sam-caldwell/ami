package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_CrossPackage_ConflictingExactVersions_StrictPromoted(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "crosspkg")
    if err := os.MkdirAll(filepath.Join(dir, "src1"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.MkdirAll(filepath.Join(dir, "src2"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    // Two packages with conflicting exact versions
    ws.Packages = append(ws.Packages, workspace.PackageEntry{Key: "lib1", Package: workspace.Package{Name: "Lib1", Version: "0.1.0", Root: "./src1", Import: []string{"github.com/acme/a == v1.2.3"}}})
    ws.Packages = append(ws.Packages, workspace.PackageEntry{Key: "lib2", Package: workspace.Package{Name: "Lib2", Version: "0.1.0", Root: "./src2", Import: []string{"github.com/acme/a == v1.3.0"}}})
    ws.Toolchain.Linter.Options = []string{"strict"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var buf bytes.Buffer
    // strict auto-enabled via workspace options
    err := runLint(&buf, dir, true, false, false)
    if err == nil { t.Fatalf("expected error due to strict promotion") }
    // ensure the diag is present
    dec := json.NewDecoder(&buf)
    var saw bool
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        if m["code"] == "E_IMPORT_CONSTRAINT_MULTI" { saw = true }
    }
    if !saw { t.Fatalf("expected E_IMPORT_CONSTRAINT_MULTI; out=%s", buf.String()) }
}

