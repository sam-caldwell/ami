package e2e

import (
    "bytes"
    "encoding/json"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

func TestE2E_AmiModList_JSON_ListsEntries(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "mod_list", "json")
    cache := filepath.Join(ws, "cache")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(filepath.Join(cache, "pkgA", "1.2.3"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(cache, "pkgA", "1.2.3", "f.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := os.WriteFile(filepath.Join(cache, "fileB"), []byte("y"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    cmd := exec.Command(bin, "mod", "list", "--json")
    cmd.Dir = ws
    absCache, _ := filepath.Abs(cache)
    cmd.Env = append(os.Environ(), "AMI_PACKAGE_CACHE="+absCache)
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil { t.Fatalf("run: %v\n%s", err, stderr.String()) }
    var res struct{ Entries []struct{ Name, Version string } }
    if err := json.Unmarshal(stdout.Bytes(), &res); err != nil { t.Fatalf("json: %v; out=%s", err, stdout.String()) }
    if len(res.Entries) < 2 { t.Fatalf("expected at least 2 entries; got %d", len(res.Entries)) }
}

func TestE2E_AmiModList_Human_EmptyOK(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "mod_list", "human_empty")
    cache := filepath.Join(ws, "cache")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(cache, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }

    cmd := exec.Command(bin, "mod", "list")
    cmd.Dir = ws
    absCache, _ := filepath.Abs(cache)
    cmd.Env = append(os.Environ(), "AMI_PACKAGE_CACHE="+absCache)
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil { t.Fatalf("run: %v\n%s", err, stderr.String()) }
}

