package main

import (
    "bytes"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestModClean_TextOutput_DefaultHomePath(t *testing.T) {
    // Use a temp HOME so we don't touch the real user dir
    home := t.TempDir()
    oldHome := os.Getenv("HOME")
    defer os.Setenv("HOME", oldHome)
    _ = os.Setenv("HOME", home)
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Unsetenv("AMI_PACKAGE_CACHE")

    var buf bytes.Buffer
    if err := runModClean(&buf, false); err != nil {
        t.Fatalf("runModClean: %v", err)
    }
    out := buf.String()
    if !strings.Contains(out, "cleaned: ") { t.Fatalf("expected human output; got: %q", out) }
    // Ensure directory exists under HOME
    target := filepath.Join(home, ".ami", "pkg")
    if st, err := os.Stat(target); err != nil || !st.IsDir() { t.Fatalf("expected %s to exist; st=%v err=%v", target, st, err) }
}

