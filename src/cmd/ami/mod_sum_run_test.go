package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "testing"
)

func TestModSum_MissingFile_JSON(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_sum", "missing")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    var buf bytes.Buffer
    err := runModSum(&buf, dir, true)
    if err == nil { t.Fatalf("expected error for missing ami.sum") }
    var res modSumResult
    if e := json.Unmarshal(buf.Bytes(), &res); e != nil { t.Fatalf("json: %v; out=%s", e, buf.String()) }
    if res.Ok { t.Fatalf("expected ok=false") }
}

func TestModSum_InvalidJSON(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_sum", "invalid")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "ami.sum"), []byte("{"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var buf bytes.Buffer
    if err := runModSum(&buf, dir, true); err == nil { t.Fatalf("expected error for invalid JSON") }
}

func TestModSum_BadSchema(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_sum", "bad_schema")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := []byte(`{"schema":"wrong","packages":[]}`)
    if err := os.WriteFile(filepath.Join(dir, "ami.sum"), content, 0o644); err != nil { t.Fatalf("write: %v", err) }
    var buf bytes.Buffer
    if err := runModSum(&buf, dir, true); err == nil { t.Fatalf("expected error for wrong schema") }
}

func TestModSum_Happy_Minimal(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_sum", "happy")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := []byte(`{"schema":"ami.sum/v1","packages":{}}`)
    if err := os.WriteFile(filepath.Join(dir, "ami.sum"), content, 0o644); err != nil { t.Fatalf("write: %v", err) }
    var buf bytes.Buffer
    if err := runModSum(&buf, dir, true); err != nil { t.Fatalf("runModSum: %v", err) }
    var res modSumResult
    if e := json.Unmarshal(buf.Bytes(), &res); e != nil { t.Fatalf("json: %v; out=%s", e, buf.String()) }
    if !res.Ok || res.PackagesSeen != 0 { t.Fatalf("unexpected result: %+v", res) }
}

func TestModSum_IntegrityMismatch(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_sum", "mismatch")
    cache := filepath.Join(dir, "cache")
    if err := os.MkdirAll(filepath.Join(cache, "alpha", "1.0.0"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(cache, "alpha", "1.0.0", "file.txt"), []byte("hello"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    // Point env at test cache
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", cache)
    // Wrong hash
    sum := []byte(`{"schema":"ami.sum/v1","packages":[{"name":"alpha","version":"1.0.0","sha256":"000000"}]}`)
    if err := os.WriteFile(filepath.Join(dir, "ami.sum"), sum, 0o644); err != nil { t.Fatalf("write sum: %v", err) }
    var buf bytes.Buffer
    err := runModSum(&buf, dir, true)
    if err == nil { t.Fatalf("expected integrity error") }
}

func TestModSum_IntegrityMatch(t *testing.T) {
    dir := filepath.Join("build", "test", "mod_sum", "match")
    cache := filepath.Join(dir, "cache")
    p := filepath.Join(cache, "beta", "2.0.0")
    if err := os.MkdirAll(p, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(p, "a.txt"), []byte("A"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := os.WriteFile(filepath.Join(p, "b.txt"), []byte("B"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", cache)
    // Compute expected hash using same logic as implementation
    got, err := hashDir(p)
    if err != nil { t.Fatalf("hashDir: %v", err) }
    sum := []byte(fmt.Sprintf(`{"schema":"ami.sum/v1","packages":[{"name":"beta","version":"2.0.0","sha256":"%s"}]}`, got))
    if err := os.WriteFile(filepath.Join(dir, "ami.sum"), sum, 0o644); err != nil { t.Fatalf("write sum: %v", err) }
    var buf bytes.Buffer
    if err := runModSum(&buf, dir, true); err != nil { t.Fatalf("runModSum: %v", err) }
    var res modSumResult
    if e := json.Unmarshal(buf.Bytes(), &res); e != nil { t.Fatalf("json: %v; out=%s", e, buf.String()) }
    if !res.Ok || res.PackagesSeen != 1 { t.Fatalf("unexpected result: %+v", res) }
}
