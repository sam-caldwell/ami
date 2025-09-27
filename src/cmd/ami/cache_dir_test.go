package main

import (
    "os"
    "path/filepath"
    "testing"
)

// Test that when AMI_PACKAGE_CACHE is unset, we create ${HOME}/.ami/pkg
func TestEnsurePackageCache_DefaultHome(t *testing.T) {
    // Use a temp HOME to avoid touching the real user directory
    tmp, err := os.MkdirTemp("", "ami-home-")
    if err != nil { t.Fatalf("temp dir: %v", err) }
    defer os.RemoveAll(tmp)

    oldHome := os.Getenv("HOME")
    oldCache := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("HOME", oldHome)
    defer os.Setenv("AMI_PACKAGE_CACHE", oldCache)

    _ = os.Setenv("HOME", tmp)
    _ = os.Unsetenv("AMI_PACKAGE_CACHE")

    if err := ensurePackageCache(); err != nil {
        t.Fatalf("ensurePackageCache: %v", err)
    }
    want := filepath.Join(tmp, ".ami", "pkg")
    st, err := os.Stat(want)
    if err != nil {
        t.Fatalf("stat %s: %v", want, err)
    }
    if !st.IsDir() { t.Fatalf("%s not a directory", want) }
}

// Test that when AMI_PACKAGE_CACHE is set to a non-existent path, it is created.
func TestEnsurePackageCache_EnvOverride(t *testing.T) {
    tmp, err := os.MkdirTemp("", "ami-cache-")
    if err != nil { t.Fatalf("temp dir: %v", err) }
    defer os.RemoveAll(tmp)

    target := filepath.Join(tmp, "custom-cache")
    oldHome := os.Getenv("HOME")
    _ = os.Unsetenv("HOME") // ensure HOME not used
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    defer os.Setenv("HOME", oldHome)
    _ = os.Setenv("AMI_PACKAGE_CACHE", target)

    // Remove if it somehow exists
    _ = os.RemoveAll(target)

    if err := ensurePackageCache(); err != nil {
        t.Fatalf("ensurePackageCache: %v", err)
    }
    st, err := os.Stat(target)
    if err != nil { t.Fatalf("stat target: %v", err) }
    if !st.IsDir() { t.Fatalf("target not a directory") }
}
