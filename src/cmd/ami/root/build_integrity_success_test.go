package root_test

import (
    "encoding/json"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"

    git "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/go-git/go-git/v5/plumbing/object"

    ammod "github.com/sam-caldwell/ami/src/ami/mod"
)

// Creates a minimal local git repo with a lightweight tag.
func makeTaggedRepo2(t *testing.T, dir, tag string) {
    t.Helper()
    r, err := git.PlainInit(dir, false)
    if err != nil { t.Fatalf("init repo: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    w, _ := r.Worktree()
    _, _ = w.Add("f.txt")
    h, err := w.Commit("c", &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@x", When: time.Now()}})
    if err != nil { t.Fatalf("commit: %v", err) }
    if err := r.Storer.SetReference(plumbing.NewHashReference(plumbing.NewTagReferenceName(tag), h)); err != nil {
        t.Fatalf("tag: %v", err)
    }
}

func TestBuild_IntegritySuccess_ExitCode0(t *testing.T) {
    // HOME for cache
    home := t.TempDir()
    t.Setenv("HOME", home)
    cache := filepath.Join(home, ".ami", "pkg")
    if err := os.MkdirAll(cache, 0o755); err != nil { t.Fatalf("mkdir cache: %v", err) }

    // Workspace
    ws := t.TempDir()
    wsContent := `version: 1.0.0
toolchain:
  compiler:
    concurrency: NUM_CPU
    target: ./build
    env: []
  linker: {}
  linter: {}
packages:
  - main:
      version: 0.0.1
      root: ./src
      import: []
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsContent), 0o644); err != nil { t.Fatalf("write workspace: %v", err) }

    version := "v1.0.0"
    entry := filepath.Join(cache, "repo@"+version)
    if err := os.MkdirAll(entry, 0o755); err != nil { t.Fatalf("mkdir entry: %v", err) }
    makeTaggedRepo2(t, entry, version)

    // Compute correct digest and write ami.sum
    d, err := ammod.CommitDigestForCLI(entry, version)
    if err != nil { t.Fatalf("commit digest: %v", err) }
    sum := map[string]any{
        "schema": "ami.sum/v1",
        "packages": map[string]map[string]string{
            "example/repo": {version: d},
        },
    }
    b, _ := json.Marshal(sum)
    if err := os.WriteFile(filepath.Join(ws, "ami.sum"), b, 0o644); err != nil { t.Fatalf("write ami.sum: %v", err) }

    // Spawn subprocess to run the build and expect success (exit 0)
    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiBuild")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI=1", "HOME="+home)
    cmd.Dir = ws
    if err := cmd.Run(); err != nil {
        if ee, ok := err.(*exec.ExitError); ok {
            t.Fatalf("unexpected non-zero exit code: %d", ee.ExitCode())
        }
        t.Fatalf("unexpected error running helper: %v", err)
    }
    // Success: optionally ensure manifest exists
    if _, err := os.Stat(filepath.Join(ws, "ami.manifest")); err != nil {
        t.Fatalf("expected ami.manifest to be written: %v", err)
    }
}

