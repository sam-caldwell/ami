package root_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
	testutil "github.com/sam-caldwell/ami/src/internal/testutil"
)

func TestInit_WritesGitignoreWithBuild(t *testing.T) {
	ws, restore := testutil.ChdirToBuildTest(t)
	defer restore()

	// run init with --force to ensure git repo or skip requirement
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ami", "init", "--force"}
	_ = rootcmd.Execute()

	gi := filepath.Join(ws, ".gitignore")
	b, err := os.ReadFile(gi)
	if err != nil {
		t.Fatalf(".gitignore missing: %v", err)
	}
	if !strings.Contains(string(b), "./build") {
		t.Fatalf(".gitignore does not include ./build: %q", string(b))
	}
}

func TestInit_GitignoreNotDuplicated(t *testing.T) {
	ws, restore := testutil.ChdirToBuildTest(t)
	defer restore()

	// Pre-create .gitignore with ./build
	if err := os.WriteFile(filepath.Join(ws, ".gitignore"), []byte("./build\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Run init again
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ami", "init", "--force"}
	_ = rootcmd.Execute()

	gi := filepath.Join(ws, ".gitignore")
	b, err := os.ReadFile(gi)
	if err != nil {
		t.Fatalf(".gitignore missing: %v", err)
	}
	count := strings.Count(string(b), "./build")
	if count != 1 {
		t.Fatalf("expected single ./build entry; got %d in %q", count, string(b))
	}
}
