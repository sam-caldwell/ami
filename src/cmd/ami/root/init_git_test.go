package root_test

import (
    "os"
    "path/filepath"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
    testutil "github.com/sam-caldwell/ami/src/internal/testutil"
)

func TestInit_CreatesGitRepoWithForce(t *testing.T) {
    ws, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"ami", "init", "--force"}
    _ = rootcmd.Execute()

    if fi, err := os.Stat(filepath.Join(ws, ".git")); err != nil || !fi.IsDir() {
        t.Fatalf("expected .git directory to exist after init --force")
    }
}

func TestInit_ErrorsWhenNotGitRepoWithoutForce(t *testing.T) {
    ws, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    // Sanity: ensure no .git exists
    if _, err := os.Stat(filepath.Join(ws, ".git")); !os.IsNotExist(err) {
        t.Fatalf("precondition: expected no .git directory in temp workspace")
    }

    oldArgs := os.Args
    defer func() { os.Args = oldArgs }()
    os.Args = []string{"ami", "init"}
    _ = rootcmd.Execute()

    // Should not create .git without --force
    if _, err := os.Stat(filepath.Join(ws, ".git")); err == nil {
        t.Fatalf("did not expect .git directory to be created without --force")
    }
    // Should not create workspace files on error
    if _, err := os.Stat(filepath.Join(ws, "ami.workspace")); err == nil {
        t.Fatalf("did not expect ami.workspace to be created when not in a git repo")
    }
    if _, err := os.Stat(filepath.Join(ws, "src", "main.ami")); err == nil {
        t.Fatalf("did not expect src/main.ami to be created when not in a git repo")
    }
}
