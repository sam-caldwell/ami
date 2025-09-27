package e2e

import (
    "bytes"
    "io"
    "os/exec"
    "strings"
    "testing"
)

func TestAmiPipelineVisualize_JSON_SchemaPresent(t *testing.T) {
    bin := buildAmi(t)
    cmd := exec.Command(bin, "pipeline", "visualize", "--json")
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil {
        t.Fatalf("run: %v\nstderr=%s\nstdout=%s", err, stderr.String(), stdout.String())
    }
    if !strings.Contains(stdout.String(), "\"schema\":\"graph.v1\"") {
        t.Fatalf("missing schema in output: %s", stdout.String())
    }
}

func TestAmiPipelineVisualize_ASCII_Line(t *testing.T) {
    bin := buildAmi(t)
    cmd := exec.Command(bin, "pipeline", "visualize")
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil {
        t.Fatalf("run: %v\nstderr=%s\nstdout=%s", err, stderr.String(), stdout.String())
    }
    got := stdout.String()
    want := "[ingress] --> (worker) --> [egress]\n"
    if got != want {
        t.Fatalf("unexpected ascii output\n got: %q\nwant: %q", got, want)
    }
}

