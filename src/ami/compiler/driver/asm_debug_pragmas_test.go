package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestAsmDebug_IncludesPragmasHeader(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\n#pragma telemetry enabled=true\nfunc F(){}\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "asm", "app", "u.s")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read asm: %v", err) }
    s := string(b)
    if !strings.Contains(s, "; pragma telemetry: enabled=true") && !strings.Contains(s, "; pragma telemetry:enabled=true") {
        t.Fatalf("telemetry pragma not found in asm header: %s", s)
    }
}

