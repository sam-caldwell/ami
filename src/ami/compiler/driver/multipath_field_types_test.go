package driver

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/workspace"
)

func testCompile_MergeField_OnPrimitive_Error(t *testing.T) {
	ws := workspace.Workspace{}
	fs := &source.FileSet{}
	// Upstream A declares Event<int>; Collect uses merge.Sort("ts") which is invalid on primitive payload.
	fs.AddFile("u.ami", "package app\npipeline P(){ A type(\"Event<int>\"); A -> Collect; Collect merge.Sort(\"ts\"); egress }\n")
	pkgs := []Package{{Name: "app", Files: fs}}
	_, ds := Compile(ws, pkgs, Options{Debug: false})
	has := false
	for _, d := range ds {
		if d.Code == "E_MERGE_FIELD_ON_PRIMITIVE" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected E_MERGE_FIELD_ON_PRIMITIVE; got %+v", ds)
	}
}

func testCompile_MergeField_Unverified_Warn(t *testing.T) {
	ws := workspace.Workspace{}
	fs := &source.FileSet{}
	// No upstream type info; should warn unverified
	fs.AddFile("u.ami", "package app\npipeline P(){ A; A -> Collect; Collect merge.Key(\"id\"); egress }\n")
	pkgs := []Package{{Name: "app", Files: fs}}
	_, ds := Compile(ws, pkgs, Options{Debug: false})
	has := false
	for _, d := range ds {
		if d.Code == "W_MERGE_FIELD_UNVERIFIED" {
			has = true
		}
	}
	if !has {
		t.Fatalf("expected W_MERGE_FIELD_UNVERIFIED; got %+v", ds)
	}
}
