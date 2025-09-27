package codegen

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

// Ensure deterministic ordering by unit name and preference of .o over .s.
func TestBuildObjIndex_DeterministicOrdering_MixedUnits(t *testing.T) {
    pkg := "mix"
    dir := filepath.Join("build", "test", "objindex_order", pkg)
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Create files in non-sorted order
    must := func(p string, b []byte) { if err := os.WriteFile(p, b, 0o644); err != nil { t.Fatalf("write %s: %v", p, err) } }
    must(filepath.Join(dir, "c.s"), []byte("; c"))
    must(filepath.Join(dir, "a.s"), []byte("; a"))
    must(filepath.Join(dir, "b.o"), []byte("BO")) // b only .o
    must(filepath.Join(dir, "a.o"), []byte("AO")) // a has both, prefer .o

    idx, err := BuildObjIndex(pkg, dir)
    if err != nil { t.Fatalf("build: %v", err) }
    if len(idx.Units) != 3 { t.Fatalf("expected 3 units, got %d", len(idx.Units)) }
    // Units should be ordered by unit name: a, b, c
    if idx.Units[0].Unit != "a" || idx.Units[1].Unit != "b" || idx.Units[2].Unit != "c" {
        t.Fatalf("order: %+v", idx.Units)
    }
    // a should pick .o; b only .o; c only .s
    if idx.Units[0].Path != "a.o" || idx.Units[1].Path != "b.o" || idx.Units[2].Path != "c.s" {
        t.Fatalf("paths: %+v", idx.Units)
    }

    if err := WriteObjIndex(idx); err != nil { t.Fatalf("write: %v", err) }
    out := filepath.Join("build", "obj", pkg, "index.json")
    b, err := os.ReadFile(out)
    if err != nil { t.Fatalf("read: %v", err) }
    var got ObjIndex
    if err := json.Unmarshal(b, &got); err != nil { t.Fatalf("json: %v", err) }
    if len(got.Units) != 3 || got.Units[0].Path != "a.o" || got.Units[1].Path != "b.o" || got.Units[2].Path != "c.s" {
        t.Fatalf("json paths: %+v", got.Units)
    }
}

