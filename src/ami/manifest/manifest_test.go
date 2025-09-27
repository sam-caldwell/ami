package manifest

import (
    "encoding/json"
    "os"
    "path/filepath"
    "reflect"
    "testing"
)

// Happy path: round-trip a simple manifest with schema and a couple of fields.
func TestManifest_SaveLoad_RoundTrip(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_manifest", "roundtrip")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    path := filepath.Join(dir, "ami.manifest")

    orig := Manifest{
        Schema: "ami.manifest/v1",
        Data: map[string]any{
            "workspace": ".",
            "artifacts": []any{"bin/ami"},
            "meta": map[string]any{"k": "v"},
        },
    }
    if err := orig.Save(path); err != nil { t.Fatalf("save: %v", err) }

    var got Manifest
    if err := got.Load(path); err != nil { t.Fatalf("load: %v", err) }
    if got.Schema != orig.Schema { t.Fatalf("schema mismatch: %q != %q", got.Schema, orig.Schema) }
    if !reflect.DeepEqual(got.Data, orig.Data) {
        b1, _ := json.Marshal(orig.Data)
        b2, _ := json.Marshal(got.Data)
        t.Fatalf("data mismatch:\norig=%s\n got=%s", string(b1), string(b2))
    }
}

// Sad path: invalid JSON; missing schema.
func TestManifest_Load_Errors(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_manifest", "errors")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    bad := filepath.Join(dir, "bad.json")
    if err := os.WriteFile(bad, []byte("not json"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var m Manifest
    if err := m.Load(bad); err == nil { t.Fatalf("expected error for invalid JSON") }

    // missing schema
    noSchema := filepath.Join(dir, "noschema.json")
    if err := os.WriteFile(noSchema, []byte("{}"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := m.Load(noSchema); err == nil { t.Fatalf("expected error for missing schema") }
}

