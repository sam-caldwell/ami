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

func TestUpdateFromWorkspace_LocalRepo_ExactVersion(t *testing.T) {
	// HOME for cache
	home := t.TempDir()
	t.Setenv("HOME", home)
	cacheDir := filepath.Join(home, ".ami", "pkg")
	_ = os.MkdirAll(cacheDir, 0o755)

	ws := t.TempDir()
	// Make local repo with tag v1.0.0
	repo := filepath.Join(ws, "absrepo")
	r, err := git.PlainInit(repo, false)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	w, _ := r.Worktree()
	_, _ = w.Add("f.txt")
	h, err := w.Commit("c", &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@x", When: time.Now()}})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	if err := r.Storer.SetReference(plumbing.NewHashReference(plumbing.NewTagReferenceName("v1.0.0"), h)); err != nil {
		t.Fatalf("tag: %v", err)
	}

	abs := repo // already absolute
	content := "version: 1.0.0\nproject: { name: demo, version: 0.0.1 }\n" +
		"toolchain: { compiler: { concurrency: NUM_CPU, target: ./build, env: [] }, linker: {}, linter: {} }\n" +
		"packages:\n  - main:\n      version: 0.0.1\n      root: ./src\n      import:\n        - " + abs + " v1.0.0\n"
	if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := UpdateFromWorkspace(filepath.Join(ws, "ami.workspace")); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cacheDir, "absrepo@v1.0.0")); err != nil {
		t.Fatalf("expected cache entry absrepo@v1.0.0: %v", err)
	}
}
