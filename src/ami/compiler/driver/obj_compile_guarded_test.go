package driver

import (
    "os"
    "path/filepath"
    "testing"

    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// This test verifies that non-debug builds attempt to compile .ll to .o.
// When clang is missing, it produces E_TOOLCHAIN_MISSING diagnostics.
func TestCompile_LL_to_Object_Guarded(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc F() { }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, di := Compile(ws, pkgs, Options{Debug: false})
    if _, err := llvme.FindClang(); err != nil {
        // expect a toolchain missing diagnostic
        seen := false
        for _, d := range di { if d.Code == "E_TOOLCHAIN_MISSING" { seen = true; break } }
        if !seen { t.Fatalf("expected E_TOOLCHAIN_MISSING when clang absent") }
    } else {
        // expect an object at build/obj/app/u.o
        o := filepath.Join("build", "obj", "app", "u.o")
        st, err := os.Stat(o)
        if err != nil || st.Size() == 0 { t.Fatalf("object not written or empty: %v", err) }
    }
}

