package main

import "io"

// runBuild validates the workspace and prepares build configuration.
// For this phase, it enforces toolchain.* constraints and emits diagnostics.
func runBuild(out io.Writer, dir string, jsonOut bool, verbose bool) error {
    // Thin wrapper to keep primary logic isolated for readability.
    return runBuildImpl(out, dir, jsonOut, verbose)
}

