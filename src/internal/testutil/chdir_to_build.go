package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ChdirToBuildTest changes the working directory to a fresh per-test
// directory under the repository root at ./build/test/<testname>.
// It returns the directory path and a restore function to defer.
func ChdirToBuildTest(t *testing.T) (string, func()) {
	t.Helper()
	// capture original cwd
	orig, _ := os.Getwd()

	// find repository root by walking up to a directory containing go.mod
	// (assumes tests are executed somewhere within the repo tree)
	dir := orig
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir { // reached filesystem root
			t.Fatalf("could not locate repository root (go.mod)")
		}
		dir = parent
	}

	// ensure base build/test directory exists
	base := filepath.Join(dir, "build", "test")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir build/test: %v", err)
	}

	// create a unique subdir per test, derived from test name
	name := strings.NewReplacer("/", "_", " ", "_", "\t", "_", "\n", "_").Replace(t.Name())
	testDir := filepath.Join(base, name)
	// start clean for each run
	_ = os.RemoveAll(testDir)
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", testDir, err)
	}
	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("chdir %s: %v", testDir, err)
	}

	return testDir, func() { _ = os.Chdir(orig) }
}
