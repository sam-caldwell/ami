package main

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Multi‑package build produces obj indexes for each package; non‑debug artifacts repeatable across runs.
func TestRunBuild_MultiPackage_Repeatability(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "multi_repeat")
    _ = os.RemoveAll(dir)
    // create package roots
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.MkdirAll(filepath.Join(dir, "util"), 0o755); err != nil { t.Fatalf("mkdir util: %v", err) }
    // write simple sources
    if err := os.WriteFile(filepath.Join(dir, "src", "a.ami"), []byte("package app\nfunc A(){}\n"), 0o644); err != nil { t.Fatalf("write app: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "util", "u.ami"), []byte("package util\nfunc U(){}\n"), 0o644); err != nil { t.Fatalf("write util: %v", err) }
    // workspace with two packages
    ws := workspace.DefaultWorkspace()
    // rename main package to app
    ws.Packages[0].Package.Name = "app"
    ws.Packages[0].Package.Root = "./src"
    // add util package
    ws.Packages = append(ws.Packages, workspace.PackageEntry{Key: "util", Package: workspace.Package{Name: "util", Version: "0.0.1", Root: "./util"}})
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }

    // run build twice (non‑verbose to test obj index repeatability)
    if err := runBuild(os.Stdout, dir, false, false); err != nil { t.Fatalf("build1: %v", err) }
    idx1App := mustRead(t, filepath.Join(dir, "build", "obj", "app", "index.json"))
    idx1Util := mustRead(t, filepath.Join(dir, "build", "obj", "util", "index.json"))
    if err := runBuild(os.Stdout, dir, false, false); err != nil { t.Fatalf("build2: %v", err) }
    idx2App := mustRead(t, filepath.Join(dir, "build", "obj", "app", "index.json"))
    idx2Util := mustRead(t, filepath.Join(dir, "build", "obj", "util", "index.json"))

    // compare parsed JSON for equality (structural)
    var a1, a2, u1, u2 map[string]any
    if err := json.Unmarshal(idx1App, &a1); err != nil { t.Fatalf("json a1: %v", err) }
    if err := json.Unmarshal(idx2App, &a2); err != nil { t.Fatalf("json a2: %v", err) }
    if err := json.Unmarshal(idx1Util, &u1); err != nil { t.Fatalf("json u1: %v", err) }
    if err := json.Unmarshal(idx2Util, &u2); err != nil { t.Fatalf("json u2: %v", err) }
    if !equalJSON(a1, a2) { t.Fatalf("app index changed across runs") }
    if !equalJSON(u1, u2) { t.Fatalf("util index changed across runs") }
}

func equalJSON(a, b map[string]any) bool {
    // naive stringify compare is fine due to deterministic encoding
    ab, _ := json.Marshal(a)
    bb, _ := json.Marshal(b)
    return string(ab) == string(bb)
}

func mustRead(t *testing.T, p string) []byte {
    t.Helper()
    b, err := os.ReadFile(p)
    if err != nil { t.Fatalf("read %s: %v", p, err) }
    return b
}
