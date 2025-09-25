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

// helper to create a repo with distinct commits per tag
func makeTaggedRepo(t *testing.T, tags []string) string {
    t.Helper()
    dir := t.TempDir()
    r, err := git.PlainInit(dir, false)
    if err != nil { t.Fatalf("init: %v", err) }
    w, _ := r.Worktree()
    for i, tag := range tags {
        fn := filepath.Join(dir, "f%02d.txt")
        _ = os.WriteFile(filepath.Join(dir, "f.txt"), []byte{byte('a' + i)}, 0o644)
        _, _ = w.Add("f.txt")
        h, err := w.Commit("c", &git.CommitOptions{Author: &object.Signature{Name:"t", Email:"t@x", When: time.Now().Add(time.Duration(i)*time.Second)}})
        if err != nil { t.Fatalf("commit: %v", err) }
        if err := r.Storer.SetReference(plumbing.NewHashReference(plumbing.NewTagReferenceName(tag), h)); err != nil { t.Fatalf("tag %s: %v", tag, err) }
        _ = fn // silence linter on format string
    }
    return dir
}

func TestResolveConstraintLocal_MajorMinorOperators(t *testing.T) {
    repo := makeTaggedRepo(t, []string{"v1.0.0", "v1.1.0", "v2.0.0", "v1.1.1-rc.1"})

    if v, err := latestTagLocal(repo); err != nil || v != "v2.0.0" {
        t.Fatalf("latestTagLocal: got %s err=%v", v, err)
    }
    if v, err := resolveConstraintLocal(repo, ">=v1.1.0"); err != nil || v != "v2.0.0" {
        t.Fatalf(">=v1.1.0 -> %s err=%v", v, err)
    }
    if v, err := resolveConstraintLocal(repo, "^v1.0.0"); err != nil || v != "v1.1.0" {
        t.Fatalf("^v1.0.0 -> %s err=%v", v, err)
    }
    if v, err := resolveConstraintLocal(repo, "~v1.1.0"); err != nil || v != "v1.1.0" {
        t.Fatalf("~v1.1.0 -> %s err=%v", v, err)
    }
    if _, err := resolveConstraintLocal(repo, ">v2.0.0"); err == nil {
        t.Fatalf("expected no match for >v2.0.0")
    }
}

