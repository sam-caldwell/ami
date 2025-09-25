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

func TestSemver_InvalidStringsRejected(t *testing.T) {
    bad := []string{"v1", "v1.2", "v1.2.3.4", "v1.2.x", "v1.2.3-", "v1.2.3+", "v.1.2.3"}
    for _, v := range bad {
        if isSemVer(v) {
            t.Fatalf("expected invalid semver to be rejected: %q", v)
        }
    }
}

func TestResolveConstraintLocal_RejectsInvalid(t *testing.T) {
    // create a repo with a valid tag
    dir := t.TempDir()
    r, err := git.PlainInit(dir, false)
    if err != nil { t.Fatalf("init: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    w, _ := r.Worktree()
    _, _ = w.Add("f.txt")
    h, err := w.Commit("c", &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@x", When: time.Now()}})
    if err != nil { t.Fatalf("commit: %v", err) }
    if err := r.Storer.SetReference(plumbing.NewHashReference(plumbing.NewTagReferenceName("v1.0.0"), h)); err != nil { t.Fatalf("tag: %v", err) }

    // invalid constraint should error
    if _, err := resolveConstraintLocal(dir, "v1.2"); err == nil {
        t.Fatalf("expected error for invalid constraint version")
    }
}

