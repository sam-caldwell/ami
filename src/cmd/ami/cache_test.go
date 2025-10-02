package main

import (
    "os"
    "testing"
)

func Test_ensurePackageCache_env(t *testing.T) {
    dir := t.TempDir()
    t.Setenv("AMI_PACKAGE_CACHE", dir)
    if err := ensurePackageCache(); err != nil { t.Fatal(err) }
    if _, err := os.Stat(dir); err != nil { t.Fatal(err) }
}

