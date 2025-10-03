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

func TestE2E_AmiModUpdate_JSON_Happy_LocalPackages(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "mod_update", "json_happy")
    cache := filepath.Join(ws, "cache")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(filepath.Join(ws, ".git"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    yaml := []byte("---\nversion: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.1.0\n      root: ./src\n      import: []\n")
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), yaml, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "src", "a.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(20*stdtime.Second))
    defer cancel()
    cmd := exec.CommandContext(ctx, bin, "mod", "update", "--json")
    cmd.Dir = ws
    absCache, _ := filepath.Abs(cache)
    cmd.Env = append(os.Environ(), "AMI_PACKAGE_CACHE="+absCache)
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil { t.Fatalf("run: %v\n%s", err, stderr.String()) }
    var res struct{ Updated []struct{ Name, Version, Path string } }
    if err := json.Unmarshal(stdout.Bytes(), &res); err != nil { t.Fatalf("json: %v; out=%s", err, stdout.String()) }
    if len(res.Updated) != 1 || res.Updated[0].Name != "app" { t.Fatalf("unexpected updated: %+v", res.Updated) }
    if _, err := os.Stat(filepath.Join(cache, "app", "0.1.0", "a.txt")); err != nil { t.Fatalf("cache content missing: %v", err) }
    if _, err := os.Stat(filepath.Join(ws, "ami.sum")); err != nil { t.Fatalf("ami.sum missing: %v", err) }
}

func TestE2E_AmiModUpdate_JSON_Sad_NoWorkspace(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "mod_update", "json_noworkspace")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(ws, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ctx2, cancel2 := context.WithTimeout(context.Background(), testutil.Timeout(20*stdtime.Second))
    defer cancel2()
    cmd := exec.CommandContext(ctx2, bin, "mod", "update", "--json")
    cmd.Dir = ws
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err == nil { t.Fatalf("expected error without workspace") }
    if stdout.Len() == 0 { t.Fatalf("expected JSON on stdout") }
}
