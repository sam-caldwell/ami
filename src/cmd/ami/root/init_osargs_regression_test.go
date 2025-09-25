package root

import (
    "os"
    "path/filepath"
    "testing"
    testutil "github.com/sam-caldwell/ami/src/internal/testutil"
)

func TestInit_WithForce_UsingOsArgs_Works(t *testing.T) {
    ws, restore := testutil.ChdirToBuildTest(t)
    defer restore()
    old := os.Args
    defer func(){ os.Args = old }()
    os.Args = []string{"ami", "init", "--force"}
    code := Execute()
    if code != 0 {
        t.Fatalf("Execute returned non-zero: %d", code)
    }
    if _, err := os.Stat(filepath.Join(ws, ".git")); err != nil {
        t.Fatalf("expected .git to be created: %v", err)
    }
}

