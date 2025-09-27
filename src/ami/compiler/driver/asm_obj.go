package driver

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// writeAsmObject emits a minimal assembly-like listing under build/obj/<pkg>/<unit>.s
// for non-verbose builds so indexes can include .s entries deterministically.
func writeAsmObject(pkg, unit string, m ir.Module) (string, error) {
    dir := filepath.Join("build", "obj", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    out := filepath.Join(dir, unit+".s")
    f, err := os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
    if err != nil { return "", err }
    defer f.Close()
    // minimal deterministic textual dump
    fmt.Fprintf(f, "; obj asm for %s/%s\n", pkg, unit)
    for _, fn := range m.Functions {
        fmt.Fprintf(f, "; function %s\n", fn.Name)
        for _, b := range fn.Blocks {
            fmt.Fprintf(f, "%s:\n", b.Name)
            for range b.Instr { fmt.Fprintln(f, "  ; instr") }
        }
    }
    return out, f.Close()
}

