package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestManifest_CrossCheckWithSumFile(t *testing.T) {
	dir := t.TempDir()
	manPath := filepath.Join(dir, "ami.manifest")
	sumPath := filepath.Join(dir, "ami.sum")

	m := &Manifest{
		Project:  Project{Name: "demo", Version: "0.0.1"},
		Packages: []Package{{Name: "example/repo", Version: "v1.0.0", Digest: "abc123", Source: "src"}},
	}
	if err := Save(manPath, m); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	sum := map[string]any{
		"schema":   "ami.sum/v1",
		"packages": map[string]map[string]string{"example/repo": {"v1.0.0": "abc123"}},
	}
	b, _ := json.Marshal(sum)
	if err := os.WriteFile(sumPath, b, 0o644); err != nil {
		t.Fatalf("write sum: %v", err)
	}

	got, err := Load(manPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := got.CrossCheckWithSumFile(sumPath); err != nil {
		t.Fatalf("cross-check: %v", err)
	}
}

func TestManifest_CrossCheckWithSumFile_Mismatch(t *testing.T) {
	dir := t.TempDir()
	manPath := filepath.Join(dir, "ami.manifest")
	sumPath := filepath.Join(dir, "ami.sum")

	m := &Manifest{
		Project:  Project{Name: "demo", Version: "0.0.1"},
		Packages: []Package{{Name: "example/repo", Version: "v1.0.0", Digest: "abc123", Source: "src"}},
	}
	if err := Save(manPath, m); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	sum := map[string]any{
		"schema":   "ami.sum/v1",
		"packages": map[string]map[string]string{"example/repo": {"v1.0.0": "deadbeef"}},
	}
	b, _ := json.Marshal(sum)
	if err := os.WriteFile(sumPath, b, 0o644); err != nil {
		t.Fatalf("write sum: %v", err)
	}

	got, err := Load(manPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := got.CrossCheckWithSumFile(sumPath); err == nil {
		t.Fatalf("expected cross-check mismatch error")
	}
}
