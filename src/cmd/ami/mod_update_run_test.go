package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestModUpdate_LocalPackages_UpdateCacheAndSum_JSON(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_update", "local")
    srcDir := filepath.Join(dir, "src")
    utilDir := filepath.Join(dir, "util")
    if err := os.MkdirAll(filepath.Join(srcDir, "pkg"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.MkdirAll(utilDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte("package app"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := os.WriteFile(filepath.Join(utilDir, "u.txt"), []byte("u"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    // Workspace with two packages
    ws := workspace.DefaultWorkspace()
    ws.Packages = workspace.PackageList{
        {Key: "main", Package: workspace.Package{Name: "app", Version: "0.1.0", Root: "./src", Import: []string{}}},
        {Key: "util", Package: workspace.Package{Name: "util", Version: "1.2.3", Root: "./util", Import: []string{}}},
    }
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir dir: %v", err) }
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save ws: %v", err) }

    cache := filepath.Join("build", "test", "mod_update", "cache")
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", cache)

    var buf bytes.Buffer
    if err := runModUpdate(&buf, dir, true); err != nil { t.Fatalf("runModUpdate: %v", err) }
    var res modUpdateResult
    if err := json.Unmarshal(buf.Bytes(), &res); err != nil { t.Fatalf("json: %v; out=%s", err, buf.String()) }
    if len(res.Updated) != 2 { t.Fatalf("expected 2 updated, got %d", len(res.Updated)) }
    // Validate cache contents
    if _, err := os.Stat(filepath.Join(cache, "app", "0.1.0", "main.ami")); err != nil { t.Fatalf("app not cached: %v", err) }
    if _, err := os.Stat(filepath.Join(cache, "util", "1.2.3", "u.txt")); err != nil { t.Fatalf("util not cached: %v", err) }
    // Validate ami.sum exists
    if _, err := os.Stat(filepath.Join(dir, "ami.sum")); err != nil { t.Fatalf("ami.sum missing: %v", err) }
}

func TestModUpdate_JSON_BadWorkspace(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_update", "badws")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // no workspace file
    var buf bytes.Buffer
    if err := runModUpdate(&buf, dir, true); err == nil {
        t.Fatalf("expected error for invalid workspace")
    }
}

