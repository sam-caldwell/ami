package e2e

import (
    "bytes"
    "encoding/json"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "context"
    stdtime "time"
    "github.com/sam-caldwell/ami/src/testutil"
)

func TestE2E_AmiModSum_JSON_Happy_FileGitFetch(t *testing.T) {
	t.Skip("Disabled until we start building remote package repositories")
	bin := buildAmi(t)
	ws := filepath.Join("build", "test", "e2e", "mod_sum", "json_happy")
	repo := filepath.Join(ws, "repo")
	cache := filepath.Join(ws, "cache")
	_ = os.RemoveAll(ws)
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	// init local git repo with tag
    run := func(args ...string) {
        ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(5*stdtime.Second))
        defer cancel()
        cmd := exec.CommandContext(ctx, "git", args...)
        cmd.Dir = repo
        if out, err := cmd.CombinedOutput(); err != nil {
            t.Fatalf("git %v: %v\n%s", args, err, out)
        }
    }
	run("init")
	if err := os.WriteFile(filepath.Join(repo, "x.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	run("add", ".")
	run("-c", "user.email=test@example.com", "-c", "user.name=test", "commit", "-m", "init")
	run("tag", "v1.0.0")

	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatalf("mkdir ws: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte("version: 1.0.0\npackages: []\n"), 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}
	// Prepare ami.sum with source pointing to repo and empty sha to force update
	absRepo, _ := filepath.Abs(repo)
	sum := []byte("{\n  \"schema\": \"ami.sum/v1\",\n  \"packages\": [\n    {\n      \"name\": \"repo\",\n      \"version\": \"v1.0.0\",\n      \"sha256\": \"\",\n      \"source\": \"file+git://" + absRepo + "\"\n    }\n  ]\n}\n")
	if err := os.WriteFile(filepath.Join(ws, "ami.sum"), sum, 0o644); err != nil {
		t.Fatalf("write sum: %v", err)
	}

    ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(20*stdtime.Second))
    defer cancel()
    cmd := exec.CommandContext(ctx, bin, "mod", "sum", "--json")
	cmd.Dir = ws
	absCache, _ := filepath.Abs(cache)
	cmd.Env = append(os.Environ(), "AMI_PACKAGE_CACHE="+absCache)
	cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("run: %v\n%s", err, stderr.String())
	}
	var res struct {
		Ok       bool
		Verified []string
	}
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		t.Fatalf("json: %v; out=%s", err, stdout.String())
	}
	if !res.Ok {
		t.Fatalf("expected ok; res=%+v\nstdout=%s", res, stdout.String())
	}
	if _, err := os.Stat(filepath.Join(cache, "repo", "v1.0.0", "x.txt")); err != nil {
		t.Fatalf("cached file missing: %v", err)
	}
}

func TestE2E_AmiModSum_JSON_Sad_NoSum(t *testing.T) {
	bin := buildAmi(t)
	ws := filepath.Join("build", "test", "e2e", "mod_sum", "json_nosum")
	_ = os.RemoveAll(ws)
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
    ctx2, cancel2 := context.WithTimeout(context.Background(), testutil.Timeout(15*stdtime.Second))
    defer cancel2()
    cmd := exec.CommandContext(ctx2, bin, "mod", "sum", "--json")
	cmd.Dir = ws
	cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err == nil {
		t.Fatalf("expected error for missing ami.sum")
	}
	if stdout.Len() == 0 {
		t.Fatalf("expected JSON on stdout")
	}
}
