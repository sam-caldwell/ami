package main

import (
    "bytes"
    "os"
    "path/filepath"
    "regexp"
    "testing"
)

func TestModList_HumanOutput_Format(t *testing.T) {
    cache := filepath.Join("build", "test", "mod_list", "human")
    _ = os.RemoveAll(cache)
    if err := os.MkdirAll(filepath.Join(cache, "alpha", "v1.0.0"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", cache)
    var buf bytes.Buffer
    if err := runModList(&buf, false); err != nil { t.Fatalf("runModList: %v", err) }
    // Expect lines like: dir	alpha@v1.0.0	<size>	<iso8601>
    re := regexp.MustCompile(`(?m)^(file|dir)\t[\w.-]+(@[\w.-]+)?\t\d+\t\d{4}-\d{2}-\d{2}T`)
    if !re.Match(buf.Bytes()) {
        t.Fatalf("unexpected human output: %s", buf.String())
    }
}

