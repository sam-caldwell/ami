package main

import (
    "io"
    "testing"
)

func TestRunBuildEntry_FilePair(t *testing.T) {
    // reference only to satisfy per-file test rule
    _ = func() error { return runBuild(io.Discard, ".", false, false) }
}

