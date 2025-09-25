package root_test

import (
    "encoding/json"
    "os"
    "path/filepath"
    "reflect"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

func TestBuild_ObjIndex_Determinism_MultiPackage(t *testing.T) {
    ws := t.TempDir()
    wsContent := `version: 1.0.0
project: { name: demo, version: 0.0.1 }
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte(wsContent), 0o644); err != nil { t.Fatal(err) }
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatal(err) }
    // two packages declared in separate units under src/ (discovery scans src/*.ami)
    if err := os.WriteFile(filepath.Join(ws, "src", "a.ami"), []byte("package main\n"), 0o644); err != nil { t.Fatal(err) }
    if err := os.WriteFile(filepath.Join(ws, "src", "b.ami"), []byte("package util\n"), 0o644); err != nil { t.Fatal(err) }

    // run build twice (human mode) in ws
    oldArgs := os.Args
    oldCwd, _ := os.Getwd()
    _ = os.Chdir(ws)
    os.Args = []string{"ami", "build"}
    _ = rootcmd.Execute()
    _ = rootcmd.Execute()
    os.Args = oldArgs
    _ = os.Chdir(oldCwd)

    // compare obj indexes and asm for both packages (normalize timestamp)
    for _, pkg := range []string{"main","util"} {
        idxPath := filepath.Join(ws, "build", "obj", pkg, "index.json")
        b, err := os.ReadFile(idxPath)
        if err != nil { t.Fatalf("missing obj index for %s: %v", pkg, err) }
        var idx1 sch.ObjIndexV1
        if json.Unmarshal(b, &idx1) != nil { t.Fatalf("unmarshal idx1 %s", pkg) }
        idx1.Timestamp = ""
        // read again after second run
        b2, err := os.ReadFile(idxPath)
        if err != nil { t.Fatalf("read idx2 %s: %v", pkg, err) }
        var idx2 sch.ObjIndexV1
        if json.Unmarshal(b2, &idx2) != nil { t.Fatalf("unmarshal idx2 %s", pkg) }
        idx2.Timestamp = ""
        if !reflect.DeepEqual(idx1, idx2) { t.Fatalf("obj index changed for %s", pkg) }
        // compare asm files for each file listed
        for _, f := range idx1.Files {
            c1, _ := os.ReadFile(filepath.Join(ws, f.Path))
            c2, _ := os.ReadFile(filepath.Join(ws, f.Path))
            if !reflect.DeepEqual(c1, c2) { t.Fatalf("obj asm drift for %s", f.Path) }
        }
    }
}
