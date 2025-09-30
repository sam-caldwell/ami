package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_CrossPkg_LocalImport_MissingPath_JSON(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "crosspkg_missing")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src", "main"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    // main package imports a missing local path
    ws.Packages[0].Package.Root = "./src/main"
    ws.Packages[0].Package.Import = []string{"./lib2"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "main", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var out bytes.Buffer
    err := runBuild(&out, dir, true, false)
    if err == nil { t.Fatalf("expected error due to missing local import path") }
    // Expect E_IMPORT_LOCAL_MISSING in JSON
    s := bufio.NewScanner(bytes.NewReader(out.Bytes()))
    found := false
    for s.Scan() {
        var m map[string]any
        if json.Unmarshal(s.Bytes(), &m) != nil { continue }
        if m["code"] == "E_IMPORT_LOCAL_MISSING" { found = true; break }
    }
    if !found { t.Fatalf("expected E_IMPORT_LOCAL_MISSING; out=\n%s", out.String()) }
}

func TestRunBuild_CrossPkg_LocalImport_Undeclared_JSON(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "crosspkg_undeclared")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src", "main"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Create lib2 folder but do not declare it as a package in workspace
    if err := os.MkdirAll(filepath.Join(dir, "lib2"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src/main"
    ws.Packages[0].Package.Import = []string{"./lib2"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "main", "u.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var out bytes.Buffer
    err := runBuild(&out, dir, true, false)
    if err == nil { t.Fatalf("expected error due to undeclared local import") }
    s := bufio.NewScanner(bytes.NewReader(out.Bytes()))
    found := false
    for s.Scan() {
        var m map[string]any
        if json.Unmarshal(s.Bytes(), &m) != nil { continue }
        if m["code"] == "E_IMPORT_LOCAL_UNDECLARED" { found = true; break }
    }
    if !found { t.Fatalf("expected E_IMPORT_LOCAL_UNDECLARED; out=\n%s", out.String()) }
}

