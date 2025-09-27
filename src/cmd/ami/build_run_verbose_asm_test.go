package main

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestRunBuild_Verbose_WritesNonEmptyASM(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "verbose_asm")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }

    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    src := "package app\nfunc F() (int) { var x int; x = 1; return x }\n"
    if err := os.WriteFile(filepath.Join(dir, "src", "unit.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    if err := runBuild(os.Stdout, dir, false, true); err != nil { t.Fatalf("runBuild: %v", err) }

    asmDir := filepath.Join(dir, "build", "debug", "asm", ws.Packages[0].Package.Name)
    matches, _ := filepath.Glob(filepath.Join(asmDir, "*.s"))
    if len(matches) == 0 { t.Fatalf("expected ASM files under %s", asmDir) }
    for _, p := range matches {
        b, err := os.ReadFile(p); if err != nil { t.Fatalf("read %s: %v", p, err) }
        if len(b) == 0 { t.Fatalf("asm file is empty: %s", p) }
    }
}

