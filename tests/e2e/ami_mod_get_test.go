package e2e

import (
    "context"
    "bytes"
    "encoding/json"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    stdtime "time"
    "github.com/sam-caldwell/ami/src/testutil"
)

func TestE2E_AmiModGet_JSON_LocalPath_Happy(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "mod_get", "json_happy")
    cache := filepath.Join(ws, "cache")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(filepath.Join(ws, ".git"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // workspace with util package
    wsYaml := []byte("---\nversion: 1.0.0\npackages:\n  - util:\n      name: util\n      version: 1.2.3\n      root: ./util\n      import: []\n")
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), wsYaml, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "util"), 0o755); err != nil { t.Fatalf("mkdir util: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "util", "a.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(20*stdtime.Second))
    defer cancel()
    cmd := exec.CommandContext(ctx, bin, "mod", "get", "./util", "--json")
    cmd.Dir = ws
    absCache, _ := filepath.Abs(cache)
    cmd.Env = append(os.Environ(), "AMI_PACKAGE_CACHE="+absCache)
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil { t.Fatalf("run: %v\n%s", err, stderr.String()) }
    var res struct{ Name, Version, Path, Message string }
    if err := json.Unmarshal(stdout.Bytes(), &res); err != nil { t.Fatalf("json: %v; out=%s", err, stdout.String()) }
    if res.Name != "util" || res.Version != "1.2.3" || res.Message != "ok" { t.Fatalf("unexpected: %+v", res) }
    if _, err := os.Stat(filepath.Join(cache, "util", "1.2.3", "a.txt")); err != nil { t.Fatalf("cached file missing: %v", err) }
    if _, err := os.Stat(filepath.Join(ws, "ami.sum")); err != nil { t.Fatalf("ami.sum missing: %v", err) }
}

func TestE2E_AmiModGet_JSON_Sad_OutsideWorkspace(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "mod_get", "json_sad_outside")
    cache := filepath.Join(ws, "cache")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(ws, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte("version: 1.0.0\npackages: []\n"), 0o644); err != nil { t.Fatalf("write ws: %v", err) }

    ctx2, cancel2 := context.WithTimeout(context.Background(), testutil.Timeout(20*stdtime.Second))
    defer cancel2()
    cmd := exec.CommandContext(ctx2, bin, "mod", "get", "../other", "--json")
    cmd.Dir = ws
    absCache, _ := filepath.Abs(cache)
    cmd.Env = append(os.Environ(), "AMI_PACKAGE_CACHE="+absCache)
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err == nil { t.Fatalf("expected error for outside path") }
    if stdout.Len() == 0 { t.Fatalf("expected JSON on stdout; got empty") }
}
