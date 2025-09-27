package workspace

import (
    "os"
    "path/filepath"
    "testing"
)

func TestUpdateManifestEntryFromCache_AndIntegrityCrossCheck(t *testing.T) {
    cache := filepath.Join("build", "test", "sum_utils", "cache")
    _ = os.RemoveAll(cache)
    if err := os.MkdirAll(filepath.Join(cache, "modA", "1.0.0"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(cache, "modA", "1.0.0", "x.txt"), []byte("hi"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    var m Manifest
    if err := UpdateManifestEntryFromCache(&m, cache, "modA", "1.0.0"); err != nil { t.Fatalf("update: %v", err) }
    if !m.Has("modA", "1.0.0") { t.Fatalf("expected manifest to have modA@1.0.0") }

    reqs := []Requirement{{Name: "modA", Constraint: Constraint{Op: OpExact, Version: "1.0.0"}}}
    miss, mis, err := CrossCheckRequirementsIntegrity(&m, reqs)
    if err != nil { t.Fatalf("integrity: %v", err) }
    if len(miss) != 0 || len(mis) != 0 { t.Fatalf("unexpected issues: miss=%v mis=%v", miss, mis) }
}

