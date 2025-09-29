package llvm

import (
    "errors"
    "os"
    "os/exec"
)

// FindClang returns the path to clang. It checks:
// - AMI_CLANG or CLANG env vars
// - common Homebrew LLVM locations
// - system PATH via exec.LookPath("clang")
func FindClang() (string, error) {
    // env overrides
    if p := os.Getenv("AMI_CLANG"); p != "" { return p, nil }
    if p := os.Getenv("CLANG"); p != "" { return p, nil }
    // common locations (Homebrew/macOS)
    candidates := []string{
        "/opt/homebrew/opt/llvm/bin/clang",
        "/usr/local/opt/llvm/bin/clang",
        "/usr/bin/clang",
    }
    for _, c := range candidates {
        if st, err := os.Stat(c); err == nil && !st.IsDir() { return c, nil }
    }
    // PATH lookup
    if p, err := exec.LookPath("clang"); err == nil { return p, nil }
    return "", errors.New("clang not found; set AMI_CLANG or CLANG to its path")
}
