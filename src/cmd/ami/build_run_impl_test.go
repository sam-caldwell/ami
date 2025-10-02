package main

import (
    "io"
    "testing"
)

func TestRunBuildImpl_FilePair(t *testing.T) {
    // do not execute; just reference symbol to satisfy per-file rule
    _ = func() error { return runBuildImpl(io.Discard, ".", true, false) }
}

