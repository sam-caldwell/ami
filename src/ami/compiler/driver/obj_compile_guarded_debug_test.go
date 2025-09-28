package driver

import (
    "os"
    "path/filepath"
    "testing"

    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure that even in Debug mode, missing clang surfaces E_TOOLCHAIN_MISSING;
// when present, an object should be written under build/obj.
func TestCompile_LL_to_Object_Debug_Guarded(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("u.ami", "package app\nfunc F() { }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, di := Compile(ws, pkgs, Options{Debug: true})
    if _, err := llvme.FindClang(); err != nil {
        // expect a toolchain missing diagnostic in debug mode
        seen := false
        for _, d := range di { if d.Code == "E_TOOLCHAIN_MISSING" { seen = true; break } }
        if !seen { t.Fatalf("expected E_TOOLCHAIN_MISSING in debug mode when clang absent; diags=%+v", di) }
    } else {
        // expect an object at build/obj/app/u.o
        o := filepath.Join("build", "obj", "app", "u.o")
        st, err := os.Stat(o)
        if err != nil || st.Size() == 0 { t.Fatalf("object not written or empty: %v", err) }
    }
}

