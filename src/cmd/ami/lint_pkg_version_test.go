package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensures workspace package versions are validated as SemVer.
func TestLint_WorkspacePackageVersion_SemVerValidation(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "pkgver")
    ws := workspace.DefaultWorkspace()
    // Set invalid version
    ws.Packages[0].Package.Version = "1.2" // invalid semver
    if err := os.MkdirAll(filepath.Join(dir), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil {
        // JSON mode returns nil even with warnings; ignore error
    }

    dec := json.NewDecoder(&buf)
    var saw bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        if m["code"] == "W_PKG_VERSION_INVALID" { saw = true }
    }
    if !saw { t.Fatalf("expected W_PKG_VERSION_INVALID; out=%s", buf.String()) }
}
