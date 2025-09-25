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

// Verifies peelToCommit can follow a tag object that points to another tag object.
func TestPeelToCommit_NestedTagObjects(t *testing.T) {
	dir := t.TempDir()
	r, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	w, _ := r.Worktree()
	_, _ = w.Add("f.txt")
	h, err := w.Commit("c", &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@x", When: time.Now()}})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	// Create annotated tag t1 -> commit
	t1Obj, err := r.CreateTag("t1", h, &git.CreateTagOptions{Tagger: &object.Signature{Name: "t", Email: "t@x", When: time.Now()}, Message: "tag1"})
	if err != nil {
		t.Fatalf("tag1: %v", err)
	}
	// Reference for t1 gives us the tag object hash
	ref1, err := r.Reference(plumbing.NewTagReferenceName("t1"), true)
	if err != nil {
		t.Fatalf("ref1: %v", err)
	}
	// Create annotated tag t2 -> t1 (tag object)
	if _, err := r.CreateTag("t2", ref1.Hash(), &git.CreateTagOptions{Tagger: &object.Signature{Name: "t", Email: "t@x", When: time.Now()}, Message: "tag2"}); err != nil {
		t.Fatalf("tag2: %v", err)
	}
	// Ensure commitDigest resolves both tags to the same commit digest
	d1, err := CommitDigestForCLI(dir, "t1")
	if err != nil {
		t.Fatalf("digest t1: %v", err)
	}
	d2, err := CommitDigestForCLI(dir, "t2")
	if err != nil {
		t.Fatalf("digest t2: %v", err)
	}
	if d1 != d2 {
		t.Fatalf("expected identical digest for nested tags; t1=%s t2=%s (t1Obj=%s)", d1, d2, t1Obj.Hash().String())
	}
}
