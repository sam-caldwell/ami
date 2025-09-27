package codegen

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestBuildObjIndex_PrefersOOverS(t *testing.T) {
    pkg := "app"
    dir := filepath.Join("build", "test", "objindex_prefer", pkg)
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "u.s"), []byte("; asm\n"), 0o644); err != nil { t.Fatalf("write s: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "u.o"), []byte("OBJ"), 0o644); err != nil { t.Fatalf("write o: %v", err) }
    idx, err := BuildObjIndex(pkg, dir)
    if err != nil { t.Fatalf("build: %v", err) }
    if len(idx.Units) != 1 { t.Fatalf("expected 1 unit, got %d", len(idx.Units)) }
    if idx.Units[0].Unit != "u" || idx.Units[0].Path != "u.o" { t.Fatalf("unexpected entry: %+v", idx.Units[0]) }
    if err := WriteObjIndex(idx); err != nil { t.Fatalf("write: %v", err) }
    out := filepath.Join("build", "obj", pkg, "index.json")
    b, err := os.ReadFile(out)
    if err != nil { t.Fatalf("read: %v", err) }
    var got ObjIndex
    if err := json.Unmarshal(b, &got); err != nil { t.Fatalf("json: %v", err) }
    if len(got.Units) != 1 || got.Units[0].Path != "u.o" { t.Fatalf("unexpected json: %+v", got) }
}

