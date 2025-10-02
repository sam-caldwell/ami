package driver

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/workspace"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testStdlibMath_LLVMIntrinsics(t *testing.T) {
	ws := workspace.Workspace{}
	fs := &source.FileSet{}
	src := "package app\nimport math\nfunc F() (float64){ return math.Sqrt(4.0) }\n"
	fs.AddFile("u.ami", src)
	pkgs := []Package{{Name: "app", Files: fs}}
	_, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
	ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
	b, err := os.ReadFile(ll)
	if err != nil {
		t.Fatalf("read llvm: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "call double @llvm.sqrt.f64") {
		t.Fatalf("missing llvm.sqrt.f64 call:\n%s", s)
	}
}

func testStdlibMath_MaxMin_LLVMIntrinsics(t *testing.T) {
	ws := workspace.Workspace{}
	fs := &source.FileSet{}
	src := "package app\nimport math\nfunc F() (float64){ return math.Max(1.0, 2.0) }\n"
	fs.AddFile("u2.ami", src)
	pkgs := []Package{{Name: "app", Files: fs}}
	_, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
	ll := filepath.Join("build", "debug", "llvm", "app", "u2.ll")
	b, err := os.ReadFile(ll)
	if err != nil {
		t.Fatalf("read llvm: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "call double @llvm.maxnum.f64") {
		t.Fatalf("missing llvm.maxnum.f64 call:\n%s", s)
	}
}
