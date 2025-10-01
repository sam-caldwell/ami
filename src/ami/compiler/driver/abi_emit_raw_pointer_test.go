package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Driver-level: pointer-like param (pointer<int>) should cause backend emission error (E_LLVM_EMIT).
func TestCompile_ABI_EmitError_OnPointerParam(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("u.ami", "package app\nfunc F(p pointer<int>){}\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(ws, pkgs, Options{Debug: true})
    found := false
    for _, d := range diags { if d.Code == "E_LLVM_EMIT" { found = true; break } }
    if !found { t.Fatalf("expected E_LLVM_EMIT for pointer-like param; got %+v", diags) }
}

// Driver-level: pointer-like result (pointer<int>) should cause backend emission error (E_LLVM_EMIT).
func TestCompile_ABI_EmitError_OnPointerResult(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("u.ami", "package app\nfunc F() (pointer<int>) { return }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(ws, pkgs, Options{Debug: true})
    found := false
    for _, d := range diags { if d.Code == "E_LLVM_EMIT" { found = true; break } }
    if !found { t.Fatalf("expected E_LLVM_EMIT for pointer-like result; got %+v", diags) }
}

