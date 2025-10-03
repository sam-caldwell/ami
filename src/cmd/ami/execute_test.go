package main

import (
    "os"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/exit"
)

func TestExecute_CoversSuccessAndErrorPaths(t *testing.T) {
    // Preserve and restore original args
    origArgs := os.Args
    t.Cleanup(func() { os.Args = origArgs })

    // Success path: no args prints help and exits 0
    os.Args = []string{"ami"}
    if code := execute(); code != 0 {
        t.Fatalf("expected exit 0, got %d", code)
    }

    // Error path: conflicting flags --json and --color trigger a user error
    os.Args = []string{"ami", "--json", "--color"}
    if code := execute(); code != exit.User.Int() {
        t.Fatalf("expected user exit %d, got %d", exit.User.Int(), code)
    }
}
