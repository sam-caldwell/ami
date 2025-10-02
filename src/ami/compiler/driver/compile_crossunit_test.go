package driver

import (
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/workspace"
)

func testCompile_CrossUnit_SignatureMismatches(t *testing.T) {
	ws := workspace.Workspace{}
	fs := &source.FileSet{}
	fs.AddFile("u1.ami", "package app\nfunc G(x int) (int){ return x }\n")
	fs.AddFile("u2.ami", "package app\nfunc F(){ var s string; G(s) }\n")                             // type mismatch
	fs.AddFile("u3.ami", "package app\nfunc H(a int, b int) (int){ return a+b }\nfunc K(){ H(1) }\n") // arity mismatch
	pkgs := []Package{{Name: "app", Files: fs}}
	_, diags := Compile(ws, pkgs, Options{Debug: false})
	var hasArg, hasArity bool
	for _, d := range diags {
		if d.Code == "E_CALL_ARG_TYPE_MISMATCH" {
			hasArg = true
		}
		if d.Code == "E_CALL_ARITY_MISMATCH" {
			hasArity = true
		}
	}
	if !hasArg || !hasArity {
		t.Fatalf("missing cross-unit call diagnostics: %v", diags)
	}
}

func testCompile_MultiPath_MismatchedUpstreams(t *testing.T) {
	ws := workspace.Workspace{}
	fs := &source.FileSet{}
	// Two upstreams with different types feed Collect
	fs.AddFile("p1.ami", "package app\npipeline P(){ ingress; A type(\"X\"); B type(\"Y\"); A -> Collect; B -> Collect; egress }\n")
	pkgs := []Package{{Name: "app", Files: fs}}
	_, diags := Compile(ws, pkgs, Options{Debug: false})
	var hasFlow bool
	for _, d := range diags {
		if d.Code == "E_EVENT_TYPE_FLOW" {
			hasFlow = true
		}
	}
	if !hasFlow {
		t.Fatalf("expected E_EVENT_TYPE_FLOW when upstream types differ: %v", diags)
	}
}
