package root_test

import (
    "bufio"
    "encoding/json"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
    "time"

    git "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/go-git/go-git/v5/plumbing/object"
)

// Ensure a summary diag with code E_INTEGRITY is emitted in JSON mode
func TestBuild_Integrity_JSONSummaryDiagEmitted(t *testing.T) {
    home := t.TempDir()
    t.Setenv("HOME", home)
    cache := filepath.Join(home, ".ami", "pkg")
    if err := os.MkdirAll(cache, 0o755); err != nil { t.Fatalf("mkdir cache: %v", err) }

    ws := t.TempDir()
    wsContent := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler:
    concurrency: 1
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

    // Prepare a mismatched cache vs ami.sum
    version := "v1.0.0"
    entry := filepath.Join(cache, "repo@"+version)
    if err := os.MkdirAll(entry, 0o755); err != nil { t.Fatalf("mkdir entry: %v", err) }
    // Tag a local repo to compute digest but put wrong digest in sum
    r, err := git.PlainInit(entry, false)
    if err != nil { t.Fatalf("init repo: %v", err) }
    if err := os.WriteFile(filepath.Join(entry, "f.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    w, _ := r.Worktree()
    _, _ = w.Add("f.txt")
    h, err := w.Commit("c", &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@x", When: time.Now()}})
    if err != nil { t.Fatalf("commit: %v", err) }
    if err := r.Storer.SetReference(plumbing.NewHashReference(plumbing.NewTagReferenceName(version), h)); err != nil { t.Fatalf("tag: %v", err) }

    sum := `{"schema":"ami.sum/v1","packages":{"example/repo":{"v1.0.0":"deadbeef"}}}`
    if err := os.WriteFile(filepath.Join(ws, "ami.sum"), []byte(sum), 0o644); err != nil { t.Fatalf("write ami.sum: %v", err) }

    cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_AmiBuildJSON")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_AMI_JSON=1", "HOME="+home)
    cmd.Dir = ws
    out, err := cmd.CombinedOutput()
    if err == nil { t.Fatalf("expected non-zero exit code; stdout=\n%s", string(out)) }

    // Expect E_INTEGRITY diag in JSON output
    type diag struct { Schema, Level, Code, Message string }
    var seen bool
    sc := bufio.NewScanner(strings.NewReader(string(out)))
    for sc.Scan() {
        var d diag
        if json.Unmarshal([]byte(sc.Text()), &d) != nil { continue }
        if d.Schema == "diag.v1" && d.Level == "error" && d.Code == "E_INTEGRITY" {
            if !strings.Contains(d.Message, "integrity violation") { t.Fatalf("unexpected message: %q", d.Message) }
            seen = true
            break
        }
    }
    if !seen { t.Fatalf("did not find summary E_INTEGRITY diag in output. stdout=\n%s", string(out)) }
}

