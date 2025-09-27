package e2e

import (
    "bytes"
    "io"
    "os/exec"
    "strings"
    "testing"
)

func TestE2E_AmiHelp_Command(t *testing.T) {
    bin := buildAmi(t)
    cmd := exec.Command(bin, "help")
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil {
        t.Fatalf("run: %v\nstderr=%s\nstdout=%s", err, stderr.String(), stdout.String())
    }
    if stderr.Len() != 0 { t.Fatalf("expected empty stderr; got: %s", stderr.String()) }
    s := stdout.String()
    if !strings.Contains(s, "AMI Help") { t.Fatalf("expected 'AMI Help' header; out=%s", s) }
    if !strings.Contains(s, "ami mod list") { t.Fatalf("expected module commands section; out=%s", s) }
}

func TestE2E_AmiRoot_HelpFlag(t *testing.T) {
    bin := buildAmi(t)
    cmd := exec.Command(bin, "--help")
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil {
        t.Fatalf("run: %v\nstderr=%s\nstdout=%s", err, stderr.String(), stdout.String())
    }
    if stderr.Len() != 0 { t.Fatalf("expected empty stderr; got: %s", stderr.String()) }
    s := stdout.String()
    if !strings.Contains(s, "Usage:") { t.Fatalf("expected Cobra usage; out=%s", s) }
    if !strings.Contains(s, "Available Commands:") { t.Fatalf("expected command list; out=%s", s) }
}

