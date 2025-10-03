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

func TestE2E_AmiRoot_Help_ShowsFlagsExamplesExitCodes(t *testing.T) {
    bin := buildAmi(t)
    ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(10*stdtime.Second))
    defer cancel()
    cmd := exec.CommandContext(ctx, bin, "--help")
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil {
        t.Fatalf("run: %v\nstderr=%s\nstdout=%s", err, stderr.String(), stdout.String())
    }
    if stderr.Len() != 0 { t.Fatalf("expected empty stderr; got: %s", stderr.String()) }
    s := stdout.String()
    // Flags
    if !strings.Contains(s, "--json") || !strings.Contains(s, "--verbose") {
        t.Fatalf("expected global flags in help; out=%s", s)
    }
    // Examples
    if !strings.Contains(s, "Examples:") { t.Fatalf("expected Examples section; out=%s", s) }
    // Exit Codes
    if !strings.Contains(s, "Exit Codes:") || !strings.Contains(s, "2 User error") {
        t.Fatalf("expected Exit Codes section; out=%s", s)
    }
}
