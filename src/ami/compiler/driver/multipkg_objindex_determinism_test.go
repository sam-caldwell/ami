package driver

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestMultiPackage_ObjIndex_DeterministicAcrossRuns(t *testing.T) {
    _ = os.RemoveAll(filepath.Join("build", "obj"))
    ws := workspace.Workspace{}
    // pkg a
    fsa := &source.FileSet{}
    fsa.AddFile("a.ami", "package a\nfunc F(){}\n")
    // pkg b
    fsb := &source.FileSet{}
    fsb.AddFile("b.ami", "package b\nfunc G(){}\n")
    pkgs := []Package{{Name: "a", Files: fsa}, {Name: "b", Files: fsb}}
    // run 1
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    ia1 := mustRead(t, filepath.Join("build", "obj", "a", "index.json"))
    ib1 := mustRead(t, filepath.Join("build", "obj", "b", "index.json"))
    // run 2
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    ia2 := mustRead(t, filepath.Join("build", "obj", "a", "index.json"))
    ib2 := mustRead(t, filepath.Join("build", "obj", "b", "index.json"))
    if string(ia1) != string(ia2) { t.Fatalf("pkg a index not deterministic") }
    if string(ib1) != string(ib2) { t.Fatalf("pkg b index not deterministic") }
}

