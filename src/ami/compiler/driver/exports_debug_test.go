package driver

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestCompile_ExportsDebug_WritesFunctionsList(t *testing.T) {
    dir := filepath.Join("build", "test", "driver_exports")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // minimal code: public function Foo gets exported
    code := "package app\nfunc Foo(){}\nfunc bar(){}\n"
    var fs source.FileSet
    fs.AddFile(filepath.Join(dir, "u.ami"), code)
    ws := workspaceDefault()
    pkgs := []Package{{Name: "app", Files: &fs}}
    oldwd, _ := os.Getwd()
    _ = os.Chdir(dir)
    Compile(ws, pkgs, Options{Debug: true})
    _ = os.Chdir(oldwd)
    p := filepath.Join(dir, "build", "debug", "link", "app", "u.exports.json")
    if st, err := os.Stat(p); err != nil || st.Size() == 0 { t.Fatalf("exports missing: %v st=%v", err, st) }
}

// workspaceDefault mirrors DefaultWorkspace but keeps this test isolated.
func workspaceDefault() (w workspace.Workspace) {
    w.Version = "1.0.0"
    w.Toolchain.Compiler.Target = "./build"
    w.Toolchain.Compiler.Env = []string{"darwin/arm64"}
    return
}
