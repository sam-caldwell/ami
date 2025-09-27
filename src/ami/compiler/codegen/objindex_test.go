package codegen

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestBuildAndWriteObjIndex(t *testing.T) {
    pkg := "app"
    dir := filepath.Join("build", "test", "objindex", pkg)
    asm := filepath.Join(dir, "unit.s")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(asm, []byte("; asm\nmov x0, x0\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }

    idx, err := BuildObjIndex(pkg, dir)
    if err != nil { t.Fatalf("BuildObjIndex: %v", err) }
    if idx.Schema != "objindex.v1" || idx.Package != pkg { t.Fatalf("bad hdr: %+v", idx) }
    if len(idx.Units) != 1 || idx.Units[0].Unit != "unit" { t.Fatalf("units: %+v", idx.Units) }

    if err := WriteObjIndex(idx); err != nil { t.Fatalf("WriteObjIndex: %v", err) }
    out := filepath.Join("build", "obj", pkg, "index.json")
    b, err := os.ReadFile(out)
    if err != nil { t.Fatalf("read index: %v", err) }
    var got ObjIndex
    if err := json.Unmarshal(b, &got); err != nil { t.Fatalf("json: %v", err) }
    if got.Package != pkg || len(got.Units) != 1 { t.Fatalf("unexpected: %+v", got) }
}

