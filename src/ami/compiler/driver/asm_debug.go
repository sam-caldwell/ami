package driver

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// writeAsmDebug emits a simple human-readable assembly-like listing per unit.
func writeAsmDebug(pkg, unit string, m ir.Module) (string, error) {
    dir := filepath.Join("build", "debug", "asm", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    out := filepath.Join(dir, unit+".s")
    f, err := os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
    if err != nil { return "", err }
    defer f.Close()
    // minimal textual dump
    fmt.Fprintf(f, "; asm listing for %s/%s\n", pkg, unit)
    for _, fn := range m.Functions {
        fmt.Fprintf(f, "; function %s\n", fn.Name)
        for _, b := range fn.Blocks {
            fmt.Fprintf(f, "%s:\n", b.Name)
            for _, ins := range b.Instr {
                _ = ins // we keep listing minimal for now
                fmt.Fprintf(f, "  ; instr\n")
            }
        }
    }
    return out, f.Close()
}

