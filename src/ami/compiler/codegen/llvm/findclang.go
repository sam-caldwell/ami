package llvm

import "os/exec"

// FindClang returns the path to clang if available on PATH.
func FindClang() (string, error) { return exec.LookPath("clang") }

