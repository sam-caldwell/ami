package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify that local imports with version constraints are checked against the local package version.
func TestLint_LocalImport_ConstraintMismatch_Warns(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "constraint_mismatch")
    if err := os.MkdirAll(filepath.Join(dir, "lib"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    // Define a local package 'lib'
    ws.Packages = append(ws.Packages, workspace.PackageEntry{Key: "lib", Package: workspace.Package{
        Name:    "LibPkg",
        Version: "0.9.0",
        Root:    "./lib",
        Import:  []string{},
    }})
    // Main imports local lib with a constraint that is not satisfied
    ws.Packages[0].Package.Import = []string{"./lib >= v1.0.0"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil {
        // in JSON mode with only warnings, error is nil; ignore
    }
    dec := json.NewDecoder(&buf)
    var saw bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        if m["code"] == "E_IMPORT_CONSTRAINT" { saw = true }
    }
    if !saw { t.Fatalf("expected E_IMPORT_CONSTRAINT for local import mismatch; out=%s", buf.String()) }
}

