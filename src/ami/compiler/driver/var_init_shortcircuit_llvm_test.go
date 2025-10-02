package driver

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	"github.com/sam-caldwell/ami/src/ami/workspace"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Non-Owned var initializer with short-circuit boolean logic
func testLower_VarInit_ShortCircuit_Bool(t *testing.T) {
	ws := workspace.Workspace{}
	fs := &source.FileSet{}
	src := "package app\nfunc F(){ var a bool; var b bool; var c bool = a || b }\n"
	fs.AddFile("u.ami", src)
	pkgs := []Package{{Name: "app", Files: fs}}
	_, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
	ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
	bts, err := os.ReadFile(ll)
	if err != nil {
		t.Fatalf("read llvm: %v", err)
	}
	s := string(bts)
	if !strings.Contains(s, "br i1 ") {
		t.Fatalf("missing cond br for var init:\n%s", s)
	}
	if !strings.Contains(s, "phi i1") {
		t.Fatalf("missing phi i1 for var init:\n%s", s)
	}
}

// Owned var initializer with short-circuit between two strings
func testLower_VarInit_ShortCircuit_Owned_String(t *testing.T) {
	ws := workspace.Workspace{}
	fs := &source.FileSet{}
	src := "package app\nfunc F(){ var a bool; var h Owned = a ? \"x\" : \"yy\" }\n"
	fs.AddFile("u2.ami", src)
	pkgs := []Package{{Name: "app", Files: fs}}
	_, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
	ll := filepath.Join("build", "debug", "llvm", "app", "u2.ll")
	bts, err := os.ReadFile(ll)
	if err != nil {
		t.Fatalf("read llvm: %v", err)
	}
	s := string(bts)
	// Expect short-circuit branch and calls to string_len + owned_new
	if !strings.Contains(s, "br i1 ") {
		t.Fatalf("missing cond br in Owned var init:\n%s", s)
	}
	if !strings.Contains(s, "declare i64 @ami_rt_string_len(ptr)") {
		t.Fatalf("missing extern for string_len:\n%s", s)
	}
	if !strings.Contains(s, "call i64 @ami_rt_string_len(") {
		t.Fatalf("missing call to string_len:\n%s", s)
	}
	if !strings.Contains(s, "call ptr @ami_rt_owned_new(") {
		t.Fatalf("missing call to owned_new:\n%s", s)
	}
}
