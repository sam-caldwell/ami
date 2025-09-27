package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestModList_DetectsVersionedSubdirs_AndSkipsUnversioned(t *testing.T) {
    cache := filepath.Join("build", "test", "mod_list", "versions")
    _ = os.RemoveAll(cache)
    if err := os.MkdirAll(filepath.Join(cache, "pkg"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Add a semver child and a non-semver child
    if err := os.MkdirAll(filepath.Join(cache, "pkg", "v1.2.3"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.MkdirAll(filepath.Join(cache, "pkg", "random"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }

    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", cache)

    var buf bytes.Buffer
    if err := runModList(&buf, true); err != nil { t.Fatalf("runModList: %v", err) }
    var res modListResult
    if err := json.Unmarshal(buf.Bytes(), &res); err != nil { t.Fatalf("json: %v; out=%s", err, buf.String()) }
    // Expect a single entry for pkg@v1.2.3; not an entry for pkg dir itself
    found := 0
    for _, e := range res.Entries {
        if e.Name == "pkg" && e.Version == "v1.2.3" { found++ }
    }
    if found != 1 { t.Fatalf("expected one pkg@v1.2.3 entry; entries=%+v", res.Entries) }
}

