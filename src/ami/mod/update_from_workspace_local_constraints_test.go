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

// makeLocalSemverRepo creates a local repo with tags and returns path and the slice of tags created.
func makeLocalSemverRepo(t *testing.T, dir string, tags []string) string {
	t.Helper()
	r, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	w, _ := r.Worktree()
	for i, tag := range tags {
		_ = os.WriteFile(filepath.Join(dir, "f.txt"), []byte{byte('a' + i)}, 0o644)
		_, _ = w.Add("f.txt")
		h, err := w.Commit("c", &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@x", When: time.Now().Add(time.Duration(i) * time.Second)}})
		if err != nil {
			t.Fatalf("commit: %v", err)
		}
		if err := r.Storer.SetReference(plumbing.NewHashReference(plumbing.NewTagReferenceName(tag), h)); err != nil {
			t.Fatalf("tag %s: %v", tag, err)
		}
	}
	return dir
}

func writeWorkspace(t *testing.T, ws string, importLine string) {
	t.Helper()
	content := "version: 1.0.0\n" +
		"project:\n  name: demo\n  version: 0.0.1\n" +
		"toolchain:\n  compiler:\n    concurrency: NUM_CPU\n    target: ./build\n    env: []\n  linker: {}\n  linter: {}\n" +
		"packages:\n  - main:\n      version: 0.0.1\n      root: ./src\n      import:\n        - " + importLine + "\n"
	if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(content), 0o644); err != nil {
		t.Fatalf("write workspace: %v", err)
	}
	_ = os.MkdirAll(filepath.Join(ws, "src"), 0o755)
}

func TestUpdateFromWorkspace_LocalRepo_SelectsByConstraint(t *testing.T) {
	// HOME cache
	home := t.TempDir()
	t.Setenv("HOME", home)
	cacheDir := filepath.Join(home, ".ami", "pkg")
	_ = os.MkdirAll(cacheDir, 0o755)

	// Workspace
	ws := t.TempDir()
	// Local repo under workspace
	repoPath := filepath.Join(ws, "lib")
	makeLocalSemverRepo(t, repoPath, []string{"v1.0.0", "v1.1.0"})

	// Case 1: ^v1.0.0 selects v1.1.0
	writeWorkspace(t, ws, "./lib ^v1.0.0")
	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()
	if err := os.Chdir(ws); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := UpdateFromWorkspace(filepath.Join(ws, "ami.workspace")); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cacheDir, "lib@v1.1.0")); err != nil {
		t.Fatalf("expected cache entry lib@v1.1.0: %v", err)
	}

	// Case 2: ~v1.0.0 selects v1.0.0
	writeWorkspace(t, ws, "./lib ~v1.0.0")
	if err := UpdateFromWorkspace(filepath.Join(ws, "ami.workspace")); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cacheDir, "lib@v1.0.0")); err != nil {
		t.Fatalf("expected cache entry lib@v1.0.0: %v", err)
	}

	// Case 3: ==latest selects v1.1.0
	writeWorkspace(t, ws, "./lib ==latest")
	if err := UpdateFromWorkspace(filepath.Join(ws, "ami.workspace")); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cacheDir, "lib@v1.1.0")); err != nil {
		t.Fatalf("expected cache entry lib@v1.1.0: %v", err)
	}
}
