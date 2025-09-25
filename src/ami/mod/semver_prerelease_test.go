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

func makeRepoWithTags(t *testing.T, dir string, tags []string) {
    t.Helper()
    r, err := git.PlainInit(dir, false)
    if err != nil { t.Fatalf("init: %v", err) }
    // base commit
    if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    w, _ := r.Worktree()
    _, _ = w.Add("f.txt")
    h, err := w.Commit("c", &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@x", When: time.Now()}})
    if err != nil { t.Fatalf("commit: %v", err) }
    // add tags in order
    for _, tag := range tags {
        if err := r.Storer.SetReference(plumbing.NewHashReference(plumbing.NewTagReferenceName(tag), h)); err != nil {
            t.Fatalf("tag %s: %v", tag, err)
        }
    }
}

func TestResolveConstraintLocal_PrereleaseFiltering(t *testing.T) {
    repo := t.TempDir()
    makeRepoWithTags(t, repo, []string{"v1.0.0-rc.1"})

    if _, err := resolveConstraintLocal(repo, "^v1.0.0"); err == nil {
        t.Fatalf("expected no matching tags when only prerelease exists and constraint has no prerelease")
    }
    v, err := resolveConstraintLocal(repo, "^v1.0.0-rc.1")
    if err != nil { t.Fatalf("resolve with prerelease: %v", err) }
    if v != "v1.0.0-rc.1" { t.Fatalf("got %s want v1.0.0-rc.1", v) }
}

func TestResolveConstraintLocal_ReleasesRankAbovePrereleases(t *testing.T) {
    repo := t.TempDir()
    // both pre and release
    makeRepoWithTags(t, repo, []string{"v1.0.0-rc.1", "v1.0.0"})

    v, err := resolveConstraintLocal(repo, "^v1.0.0")
    if err != nil { t.Fatalf("resolve: %v", err) }
    if v != "v1.0.0" { t.Fatalf("got %s want v1.0.0", v) }

    v2, err := resolveConstraintLocal(repo, "^v1.0.0-rc.1")
    if err != nil { t.Fatalf("resolve with prerelease: %v", err) }
    if v2 != "v1.0.0" { t.Fatalf("got %s want v1.0.0 (release preferred)", v2) }
}

