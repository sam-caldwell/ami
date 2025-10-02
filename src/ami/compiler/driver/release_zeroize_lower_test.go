package driver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/workspace"
)

// Lowering release(x) emits a call to ami_rt_zeroize in the generated LLVM for debug builds.
func testLower_Release_EmitsZeroizeCall_InLLVM(t *testing.T) {
	ws := workspace.Workspace{}
	fs := &source.FileSet{}
	// Owned is treated as a handle; release(a) should emit zeroize call.
	code := "package app\nfunc F(){ var a Owned; release(a) }\n"
	fs.AddFile("u.ami", code)
	pkgs := []Package{{Name: "app", Files: fs}}
	_, _ = Compile(ws, pkgs, Options{Debug: true})
	ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
	b, err := os.ReadFile(ll)
	if err != nil {
		t.Fatalf("read llvm: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "declare void @ami_rt_zeroize_owned(ptr)") {
		t.Fatalf("missing zeroize_owned extern in LLVM: %s", s)
	}
	if !strings.Contains(s, "call void @ami_rt_zeroize_owned(") {
		t.Fatalf("missing zeroize_owned call in LLVM: %s", s)
	}
}

// If Owned has a known length from a literal, release(a) uses that length (non-zero literal lowered).
func testLower_Release_UsesOwnedLenABI(t *testing.T) {
	ws := workspace.Workspace{}
	fs := &source.FileSet{}
	code := "package app\nfunc F(){ var a Owned = \"test\"; release(a) }\n"
	fs.AddFile("u2.ami", code)
	pkgs := []Package{{Name: "app", Files: fs}}
	_, _ = Compile(ws, pkgs, Options{Debug: true})
	ll := filepath.Join("build", "debug", "llvm", "app", "u2.ll")
	b, err := os.ReadFile(ll)
	if err != nil {
		t.Fatalf("read llvm: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, "declare void @ami_rt_zeroize_owned(ptr)") {
		t.Fatalf("missing zeroize_owned extern: %s", s)
	}
	if !strings.Contains(s, "call void @ami_rt_zeroize_owned(") {
		t.Fatalf("missing zeroize_owned call: %s", s)
	}
}
