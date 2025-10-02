package driver

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/workspace"
	"testing"
)

func testCompile_Capabilities_IoRequiresCapability(t *testing.T) {
	ws := workspace.Workspace{}
	fs := &source.FileSet{}
	fs.AddFile("u.ami", "package app\n#pragma trust level=trusted\npipeline P(){ ingress; io.Read(\"f\"); egress }\n")
	pkgs := []Package{{Name: "app", Files: fs}}
	_, ds := Compile(ws, pkgs, Options{Debug: false})
	found := false
	for _, d := range ds {
		if d.Code == "E_CAPABILITY_REQUIRED" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected E_CAPABILITY_REQUIRED; got %+v", ds)
	}
}

func testCompile_Capabilities_TrustUntrustedForbidsIo(t *testing.T) {
	ws := workspace.Workspace{}
	fs := &source.FileSet{}
	fs.AddFile("u.ami", "package app\n#pragma trust level=untrusted\npipeline P(){ ingress; io.Read(\"f\"); egress }\n")
	pkgs := []Package{{Name: "app", Files: fs}}
	_, ds := Compile(ws, pkgs, Options{Debug: false})
	found := false
	for _, d := range ds {
		if d.Code == "E_TRUST_VIOLATION" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected E_TRUST_VIOLATION; got %+v", ds)
	}
}
