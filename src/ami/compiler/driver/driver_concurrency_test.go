package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestDriver_PragmaConcurrency_FlowsIntoASM(t *testing.T) {
    dir := t.TempDir()
    src := "#pragma concurrency 8\npackage p\nfunc main(){}\n"
    path := filepath.Join(dir, "main.ami")
    if err := os.WriteFile(path, []byte(src), 0o644); err != nil { t.Fatal(err) }
    res, err := Compile([]string{path}, Options{})
    if err != nil { t.Fatalf("compile error: %v", err) }
    if len(res.ASM) != 1 { t.Fatalf("expected 1 asm unit; got %d", len(res.ASM)) }
    asm := res.ASM[0].Text
    if !strings.Contains(asm, "; concurrency 8") {
        t.Fatalf("asm missing concurrency: %q", asm)
    }
}

func TestDriver_PragmaScheduling_FlowsIntoASM(t *testing.T) {
    dir := t.TempDir()
    src := "#pragma scheduling fair\npackage p\nfunc main(){}\n"
    path := filepath.Join(dir, "main.ami")
    if err := os.WriteFile(path, []byte(src), 0o644); err != nil { t.Fatal(err) }
    res, err := Compile([]string{path}, Options{})
    if err != nil { t.Fatalf("compile error: %v", err) }
    if len(res.ASM) != 1 { t.Fatalf("expected 1 asm unit; got %d", len(res.ASM)) }
    asm := res.ASM[0].Text
    if !strings.Contains(asm, "; scheduling fair") {
        t.Fatalf("asm missing scheduling hint: %q", asm)
    }
}

