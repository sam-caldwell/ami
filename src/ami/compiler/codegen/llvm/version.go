package llvm

import (
    "os/exec"
    "strings"
)

// Version returns the clang --version first line (trimmed), if available.
func Version(clang string) (string, error) {
    if clang == "" { return "", ToolError{Tool: "clang", Stderr: "path empty"} }
    out, err := exec.Command(clang, "--version").CombinedOutput()
    if err != nil { return "", ToolError{Tool: "clang", Stderr: string(out)} }
    lines := strings.SplitN(string(out), "\n", 2)
    return strings.TrimSpace(lines[0]), nil
}

