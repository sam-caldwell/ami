package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure workspace.toolchain.linter.decorators_disabled disables decorators in analyzer.
func TestCompile_Decorators_Disabled_From_Workspace(t *testing.T) {
    ws := workspace.Workspace{}
    ws.Toolchain.Linter.DecoratorsDisabled = []string{"metrics"}
    code := "package app\n@metrics\nfunc F(){}\n"
    f := &source.File{Name: "a.ami", Content: code}
    fs := source.FileSet{Files: []*source.File{f}}
    pkgs := []Package{{Name: "app", Files: &fs}}
    _, diags := Compile(ws, pkgs, Options{})
    var has bool
    for _, d := range diags { if d.Code == "E_DECORATOR_DISABLED" { has = true; break } }
    if !has { t.Fatalf("expected E_DECORATOR_DISABLED when metrics is disabled; diags=%+v", diags) }
}
