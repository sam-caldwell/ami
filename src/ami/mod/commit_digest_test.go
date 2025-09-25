package mod

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestCommitDigest_AnnotatedTag(t *testing.T) {
	repoDir := t.TempDir()
	r, err := git.PlainInit(repoDir, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}
	// commit
	if err := os.WriteFile(filepath.Join(repoDir, "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	w, _ := r.Worktree()
	_, _ = w.Add("f.txt")
	h, err := w.Commit("c", &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@x", When: time.Now()}})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	// annotated tag
	if _, err := r.CreateTag("v1.2.3", h, &git.CreateTagOptions{Tagger: &object.Signature{Name: "t", Email: "t@x", When: time.Now()}, Message: "release"}); err != nil {
		t.Fatalf("create annotated tag: %v", err)
	}
	// compute digest (should not error and be deterministic)
	d1, err := CommitDigestForCLI(repoDir, "v1.2.3")
	if err != nil {
		t.Fatalf("digest error: %v", err)
	}
	d2, err := CommitDigestForCLI(repoDir, "v1.2.3")
	if err != nil {
		t.Fatalf("digest error: %v", err)
	}
	if d1 != d2 {
		t.Fatalf("digest not stable: %s vs %s", d1, d2)
	}
}
