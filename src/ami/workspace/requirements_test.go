package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func testCollectRemoteRequirements_ParsesAndSkipsLocal(t *testing.T) {
	ws := Workspace{Version: "1.0.0", Packages: PackageList{
		{Key: "main", Package: Package{
			Name: "app", Version: "0.0.1", Root: "./src",
			Import: []string{"./lib", "modA ^1.2.3", "modB >= 1.0.0", "modC 1.2.3"},
		}},
	}}
	reqs, errs := CollectRemoteRequirements(&ws)
	if len(errs) != 0 {
		t.Fatalf("unexpected errs: %v", errs)
	}
	if len(reqs) != 3 {
		t.Fatalf("want 3 reqs, got %d: %+v", len(reqs), reqs)
	}
	if reqs[0].Name != "modA" || reqs[1].Name != "modB" || reqs[2].Name != "modC" {
		t.Fatalf("names: %+v", reqs)
	}
}

func testCrossCheckRequirements_MissingAndUnsatisfied(t *testing.T) {
	dir := filepath.Join("build", "test", "reqs", "cache")
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(filepath.Join(dir, "modA", "1.2.3"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "modA", "1.2.3", "x.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	sha, err := HashDir(filepath.Join(dir, "modA", "1.2.3"))
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	m := Manifest{Schema: "ami.sum/v1"}
	m.Set("modA", "1.2.3", sha)        // satisfies ^1.2.3
	m.Set("modB", "0.9.0", "deadbeef") // does not satisfy >=1.0.0
	// modC missing entirely

	ws := Workspace{Version: "1.0.0", Packages: PackageList{
		{Key: "main", Package: Package{Name: "app", Version: "0.0.1", Root: "./src",
			Import: []string{"modA ^1.2.3", "modB >= 1.0.0", "modC 1.2.3"}}},
	}}
	reqs, _ := CollectRemoteRequirements(&ws)
	miss, unsat := CrossCheckRequirements(&m, reqs)
	if len(miss) != 1 || miss[0] != "modC" {
		t.Fatalf("missing: %v", miss)
	}
	if len(unsat) != 1 || unsat[0] != "modB" {
		t.Fatalf("unsatisfied: %v", unsat)
	}
}
