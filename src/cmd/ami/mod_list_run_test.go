package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestModList_DefaultCache_EmptyOK(t *testing.T) {
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Unsetenv("AMI_PACKAGE_CACHE")

    var buf bytes.Buffer
    if err := runModList(&buf, true); err != nil {
        t.Fatalf("runModList: %v", err)
    }
    var res modListResult
    if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
        t.Fatalf("json: %v; out=%s", err, buf.String())
    }
    if res.Path == "" {
        t.Fatalf("expected path populated")
    }
}

func TestModList_ListsEntriesSorted(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_list", "cache")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Create entries
    if err := os.MkdirAll(filepath.Join(dir, "zeta"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "alpha"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", dir)

    var buf bytes.Buffer
    if err := runModList(&buf, true); err != nil {
        t.Fatalf("runModList: %v", err)
    }
    var res modListResult
    if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
        t.Fatalf("json: %v; out=%s", err, buf.String())
    }
    if len(res.Entries) != 2 { t.Fatalf("expected 2 entries, got %d", len(res.Entries)) }
    if res.Entries[0].Name != "alpha" || res.Entries[1].Name != "zeta" {
        t.Fatalf("expected sorted entries by name, got: %+v", res.Entries)
    }
}

