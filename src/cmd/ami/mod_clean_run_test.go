package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestModClean_UsesDefaultCacheWhenEnvMissing(t *testing.T) {
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Unsetenv("AMI_PACKAGE_CACHE")

    var buf bytes.Buffer
    if err := runModClean(&buf, true); err != nil {
        t.Fatalf("runModClean: %v", err)
    }
    var res modCleanResult
    if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
        t.Fatalf("json decode: %v; out=%s", err, buf.String())
    }
    if res.Path == "" {
        t.Fatalf("expected default path in JSON")
    }
}

func TestModClean_RemovesAndRecreatesEnvPath(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_clean", "cache")
    if err := os.MkdirAll(filepath.Join(dir, "old"), 0o755); err != nil {
        t.Fatalf("mkdir: %v", err)
    }
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", dir)

    var buf bytes.Buffer
    if err := runModClean(&buf, true); err != nil {
        t.Fatalf("runModClean: %v", err)
    }
    var res modCleanResult
    if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
        t.Fatalf("json decode: %v; out=%s", err, buf.String())
    }
    if _, err := os.Stat(dir); err != nil {
        t.Fatalf("cache dir missing: %v", err)
    }
    if _, err := os.Stat(filepath.Join(dir, "old")); !os.IsNotExist(err) {
        t.Fatalf("expected old content removed; err=%v", err)
    }
}

func TestModClean_HumanOutput(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_clean", "human")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", dir)
    var buf bytes.Buffer
    if err := runModClean(&buf, false); err != nil { t.Fatalf("runModClean: %v", err) }
    if !bytes.Contains(buf.Bytes(), []byte("cleaned:")) { t.Fatalf("expected human output, got: %s", buf.String()) }
}

