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

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
)

// Helper test entrypoint that allows spawning the CLI in a subprocess.
// It only runs when GO_WANT_HELPER_AMI=1 is set.
func TestHelper_AmiBuild(t *testing.T) {
    if os.Getenv("GO_WANT_HELPER_AMI") != "1" { return }
    // Ensure CLI arguments are clean (avoid -test.* flags)
    os.Args = []string{"ami", "build"}
    code := rootcmd.Execute()
    os.Exit(code)
}

// Creates a minimal local git repo with a lightweight tag.
func makeTaggedRepo(t *testing.T, dir, tag string) {
    t.Helper()
    r, err := git.PlainInit(dir, false)
    if err != nil { t.Fatalf("init repo: %v", err) }
    // single commit
    if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    w, _ := r.Worktree()
    _, _ = w.Add("f.txt")
    h, err := w.Commit("c", &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@x", When: time.Now()}})
    if err != nil { t.Fatalf("commit: %v", err) }
    if err := r.Storer.SetReference(plumbing.NewHashReference(plumbing.NewTagReferenceName(tag), h)); err != nil {
        t.Fatalf("tag: %v", err)
    }
}

func TestBuild_IntegrityViolation_ExitCode3(t *testing.T) {
    // HOME for cache
    home := t.TempDir()
    t.Setenv("HOME", home)
    cache := filepath.Join(home, ".ami", "pkg")
    if err := os.MkdirAll(cache, 0o755); err != nil { t.Fatalf("mkdir cache: %v", err) }

    // Workspace
    ws := t.TempDir()
    // Minimal workspace file
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

    // Cache entry expected by integrity check: base(pkg)@version
    // Use pkg name "example/repo" so base is "repo"
    version := "v1.0.0"
    entry := filepath.Join(cache, "repo@"+version)
    if err := os.MkdirAll(entry, 0o755); err != nil { t.Fatalf("mkdir entry: %v", err) }
    makeTaggedRepo(t, entry, version)

    // Craft ami.sum with intentionally wrong digest for example/repo@v1.0.0
    sum := map[string]any{
        "schema": "ami.sum/v1",
        "packages": map[string]map[string]string{
            "example/repo": {version: "deadbeef"},
        },
    }
    b, _ := json.Marshal(sum)
    if err := os.WriteFile(filepath.Join(ws, "ami.sum"), b, 0o644); err != nil { t.Fatalf("write ami.sum: %v", err) }

    // Spawn subprocess to run the build and capture exit code
    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiBuild")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI=1", "HOME="+home)
    cmd.Dir = ws
    // The helper test sets its own os.Args to clean CLI args

    err := cmd.Run()
    if err == nil {
        t.Fatalf("expected build to fail with exit code 3")
    }
    if ee, ok := err.(*exec.ExitError); ok {
        if code := ee.ExitCode(); code != 3 {
            t.Fatalf("unexpected exit code: got %d want 3", code)
        }
    } else {
        t.Fatalf("unexpected error type: %T, err=%v", err, err)
    }
}
