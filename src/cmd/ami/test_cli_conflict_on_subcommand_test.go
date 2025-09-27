package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
)

// Ensure persistent flag conflict (--json and --color) is enforced on the `test` subcommand.
func TestTestCmd_CLI_FlagConflict_JSON_Color(t *testing.T) {
    dir := filepath.Join("build", "test", "cli_test_conflict")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.22\n"), 0o644); err != nil { t.Fatalf("gomod: %v", err) }
    testSrc := `package tmp
import "testing"
func TestA(t *testing.T){ }
`
    if err := os.WriteFile(filepath.Join(dir, "tmp_test.go"), []byte(testSrc), 0o644); err != nil { t.Fatalf("write: %v", err) }
    cwd, _ := os.Getwd()
    defer os.Chdir(cwd)
    _ = os.Chdir(dir)
    root := newRootCmd()
    var out bytes.Buffer
    root.SetOut(&out)
    root.SetErr(&out)
    root.SetArgs([]string{"--json", "--color", "test"})
    if err := root.Execute(); err == nil {
        t.Fatalf("expected error for --json and --color conflict")
    }
}

