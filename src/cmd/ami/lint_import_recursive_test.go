package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure linter scans local imports recursively: main -> ./lib -> ./sub
func TestLint_RecursiveLocalImports_ScansAll(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "recursive")
    // Create directories
    app := filepath.Join(dir, "app")
    lib := filepath.Join(dir, "lib")
    sub := filepath.Join(dir, "sub")
    for _, d := range []string{app, lib, sub} {
        if err := ensureDir(d); err != nil { t.Fatalf("mkdir: %v", err) }
    }
    // Add sentinels in lib and sub; app has none
    mustWrite(t, filepath.Join(lib, "lib.ami"), "UNKNOWN_IDENT\n")
    mustWrite(t, filepath.Join(sub, "sub.ami"), "UNKNOWN_IDENT\n")
    // Workspace: main imports ./lib; package lib imports ./sub
    ws := workspace.DefaultWorkspace()
    ws.Packages = append(ws.Packages, workspace.PackageEntry{Key: "lib", Package: workspace.Package{Name: "lib", Version: "0.0.1", Root: "./lib", Import: []string{"./sub"}}})
    ws.Packages[0].Package.Root = "./app"
    ws.Packages[0].Package.Import = []string{"./lib"}
    ws.Toolchain.Linter.Options = []string{"strict"} // strict: warnings -> errors
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err == nil {
        t.Fatalf("expected error in strict mode due to warnings promoted to errors")
    }
    // Decode NDJSON; ensure both lib and sub records appear
    dec := json.NewDecoder(&buf)
    var seenLib, seenSub bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        file, _ := m["file"].(string)
        code, _ := m["code"].(string)
        if code == "W_UNKNOWN_IDENT" {
            if filepath.Base(file) == "lib.ami" { seenLib = true }
            if filepath.Base(file) == "sub.ami" { seenSub = true }
        }
    }
    if !seenLib || !seenSub {
        t.Fatalf("expected unknown ident in lib and sub; got lib=%v sub=%v; out=%s", seenLib, seenSub, buf.String())
    }
}

func ensureDir(p string) error { return os.MkdirAll(p, 0o755) }

func mustWrite(t *testing.T, p, s string) {
    t.Helper()
    if err := os.WriteFile(p, []byte(s), 0o644); err != nil { t.Fatalf("write %s: %v", p, err) }
}
