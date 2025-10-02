package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sam-caldwell/ami/src/ami/workspace"
)

// Test that the verbose build plan JSON is byte-for-byte stable across repeated runs
// for the same workspace state.
func testRunBuild_Verbose_PlanDeterministicAcrossRuns(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_build", "plan_deterministic")
	_ = os.RemoveAll(dir)
	if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save: %v", err)
	}
	// minimal source to trigger compile
	if err := os.WriteFile(filepath.Join(dir, "src", "m.ami"), []byte("package app\nfunc F(){}\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// First run
	if err := runBuild(os.Stdout, dir, false, true); err != nil {
		t.Fatalf("run1: %v", err)
	}
	p := filepath.Join(dir, "build", "debug", "build.plan.json")
	b1, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read1: %v", err)
	}

	// Second run (no changes)
	if err := runBuild(os.Stdout, dir, false, true); err != nil {
		t.Fatalf("run2: %v", err)
	}
	b2, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read2: %v", err)
	}

	if !bytes.Equal(b1, b2) {
		t.Fatalf("plan not deterministic across runs:\n---1---\n%s\n---2---\n%s", string(b1), string(b2))
	}
}

// Validate plan arrays are sorted to improve determinism.
func testRunBuild_Verbose_PlanSortedCollections(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_build", "plan_sorted")
	_ = os.RemoveAll(dir)
	// Create two packages 'b' and 'a' with one unit each
	if err := os.MkdirAll(filepath.Join(dir, "a"), 0o755); err != nil {
		t.Fatalf("mkdir a: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "b"), 0o755); err != nil {
		t.Fatalf("mkdir b: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a", "a1.ami"), []byte("package a\nfunc A(){}\n"), 0o644); err != nil {
		t.Fatalf("write a1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b", "b1.ami"), []byte("package b\nfunc B(){}\n"), 0o644); err != nil {
		t.Fatalf("write b1: %v", err)
	}
	ws := workspace.Workspace{
		Version:   "1.0.0",
		Toolchain: workspace.Toolchain{Compiler: workspace.Compiler{Target: "./build", Env: workspace.DefaultWorkspace().Toolchain.Compiler.Env}},
		Packages: workspace.PackageList{
			{Key: "b", Package: workspace.Package{Name: "b", Version: "0.0.1", Root: "./b"}},
			{Key: "a", Package: workspace.Package{Name: "a", Version: "0.0.1", Root: "./a"}},
		},
	}
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save: %v", err)
	}

	if err := runBuild(os.Stdout, dir, false, true); err != nil {
		t.Fatalf("run: %v", err)
	}
	p := filepath.Join(dir, "build", "debug", "build.plan.json")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var plan struct {
		Packages []struct{ Name, Root string }
		Objects  []string `json:"objects"`
		ObjIndex []string `json:"objIndex"`
	}
	if e := json.Unmarshal(b, &plan); e != nil {
		t.Fatalf("json: %v", e)
	}
	if len(plan.Packages) != 2 {
		t.Fatalf("packages len=%d", len(plan.Packages))
	}
	if !(plan.Packages[0].Name == "a" && plan.Packages[1].Name == "b") {
		t.Fatalf("packages not sorted by name: %+v", plan.Packages)
	}
	// Objects should list 'a' entry before 'b'
	var seenA, seenB bool
	for _, s := range plan.Objects {
		if strings.Contains(s, "/a/") {
			if seenB {
				t.Fatalf("objects not sorted: %v", plan.Objects)
			}
			seenA = true
		}
		if strings.Contains(s, "/b/") {
			seenB = true
		}
	}
	if !(seenA && seenB) {
		t.Fatalf("expected objects for both a and b: %v", plan.Objects)
	}
	// ObjIndex should list a before b
	if len(plan.ObjIndex) != 2 {
		t.Fatalf("objIndex len=%d", len(plan.ObjIndex))
	}
	if !(strings.Contains(plan.ObjIndex[0], "/a/") && strings.Contains(plan.ObjIndex[1], "/b/")) {
		t.Fatalf("objIndex not sorted: %v", plan.ObjIndex)
	}
}
