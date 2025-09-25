package root

import (
    "bytes"
    "strings"
    "testing"
)

func TestHelp_Root_IncludesExamples(t *testing.T) {
    cmd := newRootCmd()
    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetErr(&buf)
    cmd.SetArgs([]string{"--help"})
    _ = cmd.Execute()
    out := buf.String()
    if !strings.Contains(out, "Examples:") {
        t.Fatalf("expected Examples section in root help; got=\n%s", out)
    }
    if !strings.Contains(out, "ami init") || !strings.Contains(out, "ami build --verbose") {
        t.Fatalf("expected example commands in root help; got=\n%s", out)
    }
}

func TestHelp_Init_IncludesExamples(t *testing.T) {
    cmd := newRootCmd()
    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetErr(&buf)
    cmd.SetArgs([]string{"init", "--help"})
    _ = cmd.Execute()
    out := buf.String()
    if !strings.Contains(out, "Examples:") || !strings.Contains(out, "ami init --force") {
        t.Fatalf("expected examples for init; got=\n%s", out)
    }
}

func TestHelp_Clean_IncludesExamples(t *testing.T) {
    cmd := newRootCmd()
    var buf bytes.Buffer
    cmd.SetOut(&buf)
    cmd.SetErr(&buf)
    cmd.SetArgs([]string{"clean", "--help"})
    _ = cmd.Execute()
    out := buf.String()
    if !strings.Contains(out, "Examples:") || !strings.Contains(out, "ami --json clean") {
        t.Fatalf("expected examples for clean; got=\n%s", out)
    }
}

