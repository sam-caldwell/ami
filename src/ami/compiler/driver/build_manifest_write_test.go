package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestWriteBuildManifest_DefaultSchemaAndContents(t *testing.T) {
    m := BuildManifest{Packages: []bmPackage{{Name: "main", Units: []bmUnit{{Unit: "u1", IR: "ir.json"}}}}}
    path, err := writeBuildManifest(m)
    if err != nil { t.Fatalf("write: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read: %v", err) }
    var got BuildManifest
    if err := json.Unmarshal(b, &got); err != nil { t.Fatalf("json: %v", err) }
    if got.Schema != "manifest.v1" { t.Fatalf("default schema expected, got %q", got.Schema) }
    if len(got.Packages) != 1 || got.Packages[0].Name != "main" || len(got.Packages[0].Units) != 1 {
        t.Fatalf("unexpected manifest: %+v", got)
    }
    // ensure file lives under build/debug
    if filepath.Base(filepath.Dir(path)) != "debug" { t.Fatalf("unexpected dir: %s", path) }
}

