package mod

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSum_Save_DeterministicOrdering(t *testing.T) {
	dir := t.TempDir()
	path1 := filepath.Join(dir, "ami1.sum")
	path2 := filepath.Join(dir, "ami2.sum")

	s := &Sum{Schema: "ami.sum/v1", Packages: map[string]map[string]string{
		"b/pkg": {"v1.0.1": "d2", "v1.0.0": "d1"},
		"a/pkg": {"v2.0.0": "x"},
	}}
	if err := saveSum(path1, s); err != nil {
		t.Fatalf("save1: %v", err)
	}

	// Different in-memory order should still produce identical file
	s2 := &Sum{Schema: "ami.sum/v1", Packages: map[string]map[string]string{
		"a/pkg": {"v2.0.0": "x"},
		"b/pkg": {"v1.0.0": "d1", "v1.0.1": "d2"},
	}}
	if err := saveSum(path2, s2); err != nil {
		t.Fatalf("save2: %v", err)
	}

	b1, _ := os.ReadFile(path1)
	b2, _ := os.ReadFile(path2)
	if string(b1) != string(b2) {
		t.Fatalf("ami.sum writes not deterministic:\n%s\n---\n%s", string(b1), string(b2))
	}
}
