package workspace

import (
    "os"
    "path/filepath"
    "testing"
)

func TestManifest_Load_ObjectForm(t *testing.T) {
    dir := filepath.Join("build", "test", "workspace_manifest", "obj")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := []byte(`{
  "schema": "ami.sum/v1",
  "packages": {
    "alpha": {"version": "v1.0.0", "sha256": "aaa"},
    "beta":  {"version": "v2.0.0", "sha256": "bbb"}
  }
}`)
    p := filepath.Join(dir, "ami.sum")
    if err := os.WriteFile(p, content, 0o644); err != nil { t.Fatalf("write: %v", err) }
    var m Manifest
    if err := m.Load(p); err != nil { t.Fatalf("Load: %v", err) }
    if m.Schema != "ami.sum/v1" { t.Fatalf("schema: %s", m.Schema) }
    if m.Packages["alpha"]["v1.0.0"] != "aaa" { t.Fatalf("alpha missing") }
    if m.Packages["beta"]["v2.0.0"] != "bbb" { t.Fatalf("beta missing") }
}

func TestManifest_Save_RoundTrip_Nested(t *testing.T) {
    dir := filepath.Join("build", "test", "workspace_manifest", "save")
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    m := Manifest{Schema: "ami.sum/v1", Packages: map[string]map[string]string{
        "alpha": {"v1.0.0": "aaa", "v1.1.0": "aab"},
        "beta":  {"v2.0.0": "bbb"},
    }}
    p := filepath.Join(dir, "ami.sum")
    if err := m.Save(p); err != nil { t.Fatalf("Save: %v", err) }
    // Reload and compare presence
    var m2 Manifest
    if err := m2.Load(p); err != nil { t.Fatalf("Load2: %v", err) }
    if m2.Packages["alpha"]["v1.0.0"] != "aaa" || m2.Packages["alpha"]["v1.1.0"] != "aab" { t.Fatalf("alpha mismatch") }
    if m2.Packages["beta"]["v2.0.0"] != "bbb" { t.Fatalf("beta mismatch") }
}

func TestManifest_Validate_UsesCacheAndHashes(t *testing.T) {
    dir := filepath.Join("build", "test", "workspace_manifest", "validate")
    cache := filepath.Join(dir, "cache")
    if err := os.MkdirAll(cache, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // create one package version with content
    p := filepath.Join(cache, "gamma", "v3.0.0")
    if err := os.MkdirAll(p, 0o755); err != nil { t.Fatalf("mkdir p: %v", err) }
    if err := os.WriteFile(filepath.Join(p, "file.txt"), []byte("hi"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", cache)
    // compute sha for manifest
    sha, err := HashDir(p)
    if err != nil { t.Fatalf("HashDir: %v", err) }
    m := Manifest{Schema: "ami.sum/v1", Packages: map[string]map[string]string{
        "gamma": {"v3.0.0": sha},    // verified
        "delta": {"v4.0.0": "xxxx"}, // missing
    }}
    v, miss, mm, err := m.Validate()
    if err != nil { t.Fatalf("Validate: %v", err) }
    if len(v) != 1 || v[0] != "gamma@v3.0.0" { t.Fatalf("verified mismatch: %+v", v) }
    if len(miss) != 1 || miss[0] != "delta@v4.0.0" { t.Fatalf("missing mismatch: %+v", miss) }
    if len(mm) != 0 { t.Fatalf("mismatched unexpected: %+v", mm) }
}

