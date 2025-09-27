package driver

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// writeAsmDebug emits a simple human-readable assembly-like listing per unit.
func writeAsmDebug(pkg, unit string, af *ast.File, m ir.Module) (string, error) {
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
    // Emit multipath pseudo-ops for pipelines
    if af != nil {
        for _, d := range af.Decls {
            if pd, ok := d.(*ast.PipelineDecl); ok {
                for _, s := range pd.Stmts {
                    if st, ok := s.(*ast.StepStmt); ok && st.Name == "Collect" {
                        // detect MultiPath and merge attributes
                        var mpArgs []string
                        var merges []mergeAttr
                        for _, at := range st.Attrs {
                            if at.Name == "edge.MultiPath" || at.Name == "MultiPath" {
                                for _, a := range at.Args { mpArgs = append(mpArgs, a.Text) }
                            }
                            if len(at.Name) >= 6 && at.Name[:6] == "merge." {
                                var margs []string
                                for _, a := range at.Args { margs = append(margs, a.Text) }
                                merges = append(merges, mergeAttr{Name: at.Name, Args: margs})
                            }
                        }
                        if len(mpArgs) > 0 {
                            fmt.Fprintf(f, "; mp_multipath args=%v\n", mpArgs)
                        }
                        for _, m := range merges {
                            fmt.Fprintf(f, "; mp_merge %s(%v)\n", m.Name, m.Args)
                        }
                    }
                }
            }
        }
    }
    return out, f.Close()
}
