package e2e

import (
    "context"
    "bytes"
    "io"
    "os/exec"
    "strings"
    "testing"
    stdtime "time"
    "github.com/sam-caldwell/ami/src/testutil"
)

func TestE2E_AmiHelp_Command(t *testing.T) {
    bin := buildAmi(t)
    ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(10*stdtime.Second))
    defer cancel()
    cmd := exec.CommandContext(ctx, bin, "help")
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
    ctx2, cancel2 := context.WithTimeout(context.Background(), testutil.Timeout(10*stdtime.Second))
    defer cancel2()
    cmd := exec.CommandContext(ctx2, bin, "--help")
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
