package main

import "testing"

// Verifies that persistent root flags conflict (--json with --color) also surfaces when running the lint subcommand.
func TestLint_Subcommand_JsonColorConflict(t *testing.T) {
    cmd := newRootCmd()
    cmd.SetArgs([]string{"--json", "--color", "lint"})
    if err := cmd.Execute(); err == nil {
        t.Fatalf("expected error for --json with --color on lint")
    }
}

