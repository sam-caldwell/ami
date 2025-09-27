package main

import (
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/exit"
)

func TestExecute_NoArgs_ReturnsSuccess(t *testing.T) {
    old := os.Args
    defer func() { os.Args = old }()
    os.Args = []string{"ami"}
    if code := execute(); code != 0 {
        t.Fatalf("expected code 0, got %d", code)
    }
}

func TestExecute_JsonColorConflict_ReturnsUserError(t *testing.T) {
    old := os.Args
    defer func() { os.Args = old }()
    os.Args = []string{"ami", "--json", "--color"}
    if code := execute(); code != exit.User.Int() {
        t.Fatalf("expected user error code, got %d", code)
    }
}

