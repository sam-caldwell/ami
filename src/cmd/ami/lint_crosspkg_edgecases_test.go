package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sam-caldwell/ami/src/ami/workspace"
)

// ^v1.2.3 ([1.2.3,2.0.0)) and >= v2.0.0 ([2.0.0,inf)) intersect at a boundary point only; since the
// upper bound is exclusive on the caret range, the intersection is empty and should produce E_IMPORT_CONSTRAINT.
func testLint_CrossPackage_RangeIntersection_EmptyAtBoundary(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_lint", "range_empty")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	ws.Toolchain.Linter.Options = []string{} // non-strict for clear warnings
	// Add a second package with a conflicting range
	ws.Packages = append(ws.Packages, workspace.PackageEntry{Key: "lib", Package: workspace.Package{
		Name:    "lib",
		Version: "0.0.1",
		Root:    "./lib",
	}})
	// Two packages import the same path with ranges that have empty intersection
	ws.Packages[0].Package.Import = []string{"github.com/x/mod ^ v1.2.3"}
	ws.Packages[1].Package.Import = []string{"github.com/x/mod >= v2.0.0"}
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save: %v", err)
	}
	var buf bytes.Buffer
	_ = runLint(&buf, dir, true, false, false)
	dec := json.NewDecoder(&buf)
	var sawConflict, sawSingle bool
	for dec.More() {
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			t.Fatalf("json: %v", err)
		}
		switch m["code"] {
		case "E_IMPORT_CONSTRAINT":
			sawConflict = true
		case "W_IMPORT_SINGLE_VERSION":
			sawSingle = true
		}
	}
	if !sawConflict {
		t.Fatalf("expected E_IMPORT_CONSTRAINT; out=%s", buf.String())
	}
	if sawSingle {
		t.Fatalf("did not expect W_IMPORT_SINGLE_VERSION alongside hard conflict; out=%s", buf.String())
	}
}

// ~v1.2.3 ([1.2.3,1.3.0)) and >= v1.2.5 ([1.2.5,inf)) overlap; no E_IMPORT_CONSTRAINT should appear.
// Because only ranges are present, we should see W_IMPORT_SINGLE_VERSION to nudge pinning in strict mode.
func testLint_CrossPackage_RangeIntersection_Overlapping(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_lint", "range_overlap")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	ws.Toolchain.Linter.Options = []string{} // non-strict
	ws.Packages = append(ws.Packages, workspace.PackageEntry{Key: "lib", Package: workspace.Package{
		Name:    "lib",
		Version: "0.0.1",
		Root:    "./lib",
	}})
	ws.Packages[0].Package.Import = []string{"github.com/x/mod ~ v1.2.3"}
	ws.Packages[1].Package.Import = []string{"github.com/x/mod >= v1.2.5"}
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save: %v", err)
	}
	var buf bytes.Buffer
	_ = runLint(&buf, dir, true, false, false)
	dec := json.NewDecoder(&buf)
	var sawConflict, sawSingle bool
	for dec.More() {
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			t.Fatalf("json: %v", err)
		}
		switch m["code"] {
		case "E_IMPORT_CONSTRAINT":
			sawConflict = true
		case "W_IMPORT_SINGLE_VERSION":
			sawSingle = true
		}
	}
	if sawConflict {
		t.Fatalf("did not expect E_IMPORT_CONSTRAINT; out=%s", buf.String())
	}
	if !sawSingle {
		t.Fatalf("expected W_IMPORT_SINGLE_VERSION when only ranges present; out=%s", buf.String())
	}
}
