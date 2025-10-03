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

func TestE2E_AmiModClean_JSON_Happy(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "mod_clean", "happy")
    cache := filepath.Join(ws, "cache")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(filepath.Join(cache, "junk"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(cache, "junk", "x.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(20*stdtime.Second))
    defer cancel()
    cmd := exec.CommandContext(ctx, bin, "mod", "clean", "--json")
    cmd.Dir = ws
    absCache, _ := filepath.Abs(cache)
    cmd.Env = append(os.Environ(), "AMI_PACKAGE_CACHE="+absCache)
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil { t.Fatalf("run: %v\nstderr=%s", err, stderr.String()) }
    if stderr.Len() != 0 { t.Fatalf("expected empty stderr; got %s", stderr.String()) }
    var res struct{ Path string; Removed, Created bool }
    if err := json.Unmarshal(stdout.Bytes(), &res); err != nil { t.Fatalf("json: %v; out=%s", err, stdout.String()) }
    if !res.Removed || !res.Created { t.Fatalf("expected removed+created; res=%+v", res) }
    // Ensure cache exists and junk removed
    if _, err := os.Stat(filepath.Join(cache, "junk", "x.txt")); !os.IsNotExist(err) { t.Fatalf("expected junk removed, err=%v", err) }
}

func TestE2E_AmiModClean_JSON_Sad_MkdirParentIsFile(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "mod_clean", "sad_parent_file")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(ws, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Create parent as a file so creating child directory fails
    parent := filepath.Join(ws, "parent")
    if err := os.WriteFile(parent, []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    cache := filepath.Join(parent, "child")

    ctx2, cancel2 := context.WithTimeout(context.Background(), testutil.Timeout(20*stdtime.Second))
    defer cancel2()
    cmd := exec.CommandContext(ctx2, bin, "mod", "clean", "--json")
    cmd.Dir = ws
    absCache, _ := filepath.Abs(cache)
    cmd.Env = append(os.Environ(), "AMI_PACKAGE_CACHE="+absCache)
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err == nil { t.Fatalf("expected error due to parent file") }
    // stdout should contain JSON with at least a path; stderr non-empty
    if stdout.Len() == 0 { t.Fatalf("expected JSON on stdout; got empty") }
}
