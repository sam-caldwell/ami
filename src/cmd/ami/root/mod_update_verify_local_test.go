package root_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	ammod "github.com/sam-caldwell/ami/src/ami/mod"
	rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
	testutil "github.com/sam-caldwell/ami/src/internal/testutil"
)

// captureStdout moved to testhelpers_test.go

func makeLocalRepo(t *testing.T, path string) (string, string) {
	t.Helper()
	r, err := git.PlainInit(path, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}
	// commit 1
	if err := os.WriteFile(filepath.Join(path, "file.txt"), []byte("one"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	w, _ := r.Worktree()
	_, _ = w.Add("file.txt")
	h1, err := w.Commit("c1", &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@x", When: time.Now()}})
	if err != nil {
		t.Fatalf("commit1: %v", err)
	}
	// lightweight tag v0.1.0
	if err := r.Storer.SetReference(plumbing.NewHashReference(plumbing.NewTagReferenceName("v0.1.0"), h1)); err != nil {
		t.Fatalf("tag1: %v", err)
	}
	// commit 2
	if err := os.WriteFile(filepath.Join(path, "file.txt"), []byte("two"), 0o644); err != nil {
		t.Fatalf("write2: %v", err)
	}
	_, _ = w.Add("file.txt")
	h2, err := w.Commit("c2", &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@x", When: time.Now().Add(time.Second)}})
	if err != nil {
		t.Fatalf("commit2: %v", err)
	}
	// lightweight tag v0.2.0
	if err := r.Storer.SetReference(plumbing.NewHashReference(plumbing.NewTagReferenceName("v0.2.0"), h2)); err != nil {
		t.Fatalf("tag2: %v", err)
	}
	return "v0.1.0", "v0.2.0"
}

// Test mod update with a local repo (no network) and verify ami.sum + cache and mod verify pass.
func TestModUpdateAndVerify_LocalRepo(t *testing.T) {
	tmp := t.TempDir()
	// HOME for cache
	t.Setenv("HOME", tmp)
	cacheDir := filepath.Join(tmp, ".ami", "pkg")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatalf("mkdir cache: %v", err)
	}

	// workspace dir under ./build/test
	ws, restore := testutil.ChdirToBuildTest(t)
	defer restore()
	// local git repo under workspace
	repoPath := filepath.Join(ws, "example")
	_, latest := makeLocalRepo(t, repoPath)

	// ami.workspace with local import ==latest
	wsContent := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
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
      import:
        - ./example ==latest
`
	if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsContent), 0o644); err != nil {
		t.Fatalf("write workspace: %v", err)
	}

	// src dir exists (not needed by mod, but keep structure tidy)
	_ = os.MkdirAll(filepath.Join(ws, "src"), 0o755)

	// already in workspace

	// Run: ami mod update
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ami", "mod", "update"}
	_ = captureStdout(t, func() { _ = rootcmd.Execute() })

	// ami.sum should exist and cache should have example@latest
	sumPath := filepath.Join(ws, "ami.sum")
	if _, err := os.Stat(sumPath); err != nil {
		t.Fatalf("ami.sum missing: %v", err)
	}

	// Verify expected cache entry exists
	if fi, err := os.Stat(filepath.Join(cacheDir, "example@"+latest)); err != nil || !fi.IsDir() {
		t.Fatalf("expected cache entry example@%s", latest)
	}

	// Compute digest directly and compare with ami.sum entry
	sum, err := ammod.LoadSumForCLI(sumPath)
	if err != nil {
		t.Fatalf("load sum: %v", err)
	}
	var found bool
	for pkg, vers := range sum.Packages {
		// key will be the repoPath string (./example). base used by verify is directory name
		if filepath.Base(pkg) == "example" {
			if d, ok := vers[latest]; ok {
				found = true
				// recompute digest from cached repo
				d2, err := ammod.CommitDigestForCLI(filepath.Join(cacheDir, "example@"+latest), latest)
				if err != nil {
					t.Fatalf("commit digest: %v", err)
				}
				if d != d2 {
					t.Fatalf("digest mismatch: sum=%s computed=%s", d, d2)
				}
			}
		}
	}
	if !found {
		t.Fatalf("did not find example@%s in ami.sum", latest)
	}

	// Run: ami mod verify (should log success)
	os.Args = []string{"ami", "mod", "verify"}
	out := captureStdout(t, func() { _ = rootcmd.Execute() })
	if !strings.Contains(out, "ami.sum verified") {
		// If JSON output accidentally enabled, try to parse for clarity
		var lines []map[string]interface{}
		sc := bufio.NewScanner(strings.NewReader(out))
		for sc.Scan() {
			var m map[string]interface{}
			_ = json.Unmarshal([]byte(sc.Text()), &m)
			if m != nil {
				lines = append(lines, m)
			}
		}
		t.Fatalf("verify did not report success. stdout=\n%s\nparsed=%v", out, lines)
	}
}
