package mod

import (
    "os"
    "path/filepath"
    "testing"
    "time"

    git "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/go-git/go-git/v5/plumbing/object"
)

func TestUpdateSum_WritesDigestForRepoTag(t *testing.T) {
    // Local repo with a tag
    repoDir := t.TempDir()
    r, err := git.PlainInit(repoDir, false)
    if err != nil { t.Fatalf("init: %v", err) }
    if err := os.WriteFile(filepath.Join(repoDir, "f.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    w, _ := r.Worktree()
    _, _ = w.Add("f.txt")
    h, err := w.Commit("c", &git.CommitOptions{Author: &object.Signature{Name:"t", Email:"t@x", When: time.Now()}})
    if err != nil { t.Fatalf("commit: %v", err) }
    if err := r.Storer.SetReference(plumbing.NewHashReference(plumbing.NewTagReferenceName("v1.0.0"), h)); err != nil {
        t.Fatalf("tag: %v", err)
    }

    // Write sum
    sumPath := filepath.Join(t.TempDir(), "ami.sum")
    dest := repoDir // we can point to repoDir for digest computation
    if err := UpdateSum(sumPath, "example/repo", "v1.0.0", dest, "v1.0.0"); err != nil {
        t.Fatalf("UpdateSum: %v", err)
    }
    // Load and verify entry exists and matches computed digest
    s, err := LoadSumForCLI(sumPath)
    if err != nil { t.Fatalf("LoadSum: %v", err) }
    if s.Packages["example/repo"] == nil { t.Fatalf("missing package entry") }
    got := s.Packages["example/repo"]["v1.0.0"]
    if got == "" { t.Fatalf("missing version entry") }
    d2, err := CommitDigestForCLI(dest, "v1.0.0")
    if err != nil { t.Fatalf("digest: %v", err) }
    if got != d2 { t.Fatalf("digest mismatch: sum=%s computed=%s", got, d2) }
}

