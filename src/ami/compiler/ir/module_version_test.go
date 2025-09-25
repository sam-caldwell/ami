package ir

import (
	astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"testing"
)

func TestIR_ToSchema_IncludesPackageVersion(t *testing.T) {
	f := &astpkg.File{Package: "p", Version: "0.0.1"}
	m := FromASTFile("p", f.Version, "u.ami", f)
	ir := m.ToSchema()
	if ir.Version != "0.0.1" {
		t.Fatalf("ir version=%q", ir.Version)
	}
}
