package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_CrossPackage_JSONAndHuman_Parity(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "crosspkg_parity")
    if err := os.MkdirAll(filepath.Join(dir, "lib1"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.MkdirAll(filepath.Join(dir, "lib2"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages = append(ws.Packages, workspace.PackageEntry{Key: "lib1", Package: workspace.Package{Name: "Lib1", Version: "0.1.0", Root: "./lib1", Import: []string{"github.com/acme/a v1.2.3"}}})
    ws.Packages = append(ws.Packages, workspace.PackageEntry{Key: "lib2", Package: workspace.Package{Name: "Lib2", Version: "0.1.0", Root: "./lib2", Import: []string{"github.com/acme/a v1.3.0"}}})
    ws.Toolchain.Linter.Options = []string{} // non-strict
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    // JSON mode: expect E_IMPORT_CONSTRAINT_MULTI
    var jb bytes.Buffer
    if err := runLint(&jb, dir, true, false, false); err != nil {
        t.Fatalf("json lint: %v\n%s", err, jb.String())
    }
    dec := json.NewDecoder(&jb)
    var saw bool
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        if m["code"] == "E_IMPORT_CONSTRAINT_MULTI" { saw = true }
    }
    if !saw { t.Fatalf("expected E_IMPORT_CONSTRAINT_MULTI in JSON") }

    // Human mode: should include code string and summary
    var hb bytes.Buffer
    if err := runLint(&hb, dir, false, false, false); err != nil {
        t.Fatalf("human lint: %v\n%s", err, hb.String())
    }
    s := hb.String()
    if !strings.Contains(s, "E_IMPORT_CONSTRAINT_MULTI") || !strings.Contains(s, "warning(s)") {
        t.Fatalf("expected human output to contain code and summary: %s", s)
    }
}

