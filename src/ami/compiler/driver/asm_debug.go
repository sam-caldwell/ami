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
    if af != nil {
        for _, pr := range af.Pragmas {
            // print a stable pragma header line
            fmt.Fprintf(f, "; pragma %s:%s %s\n", pr.Domain, pr.Key, pr.Value)
            if len(pr.Args) > 0 {
                fmt.Fprintf(f, ";  args %v\n", pr.Args)
            }
            if len(pr.Params) > 0 {
                fmt.Fprintf(f, ";  params %v\n", pr.Params)
            }
        }
    }
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
    // Emit multipath pseudo-ops and edge warnings for pipelines
    if af != nil {
        for _, d := range af.Decls {
            if pd, ok := d.(*ast.PipelineDecl); ok {
                // collect attrs by step
                stepAttrs := map[string][]ast.Attr{}
                for _, s := range pd.Stmts {
                    if st, ok := s.(*ast.StepStmt); ok { stepAttrs[st.Name] = st.Attrs }
                }
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
                    if e, ok := s.(*ast.EdgeStmt); ok {
                        // tiny buffer warnings on edge via target step attrs
                        atts := stepAttrs[e.To]
                        tiny := false
                        for _, at := range atts {
                            if at.Name == "merge.Buffer" {
                                if len(at.Args) > 0 && (at.Args[0].Text == "0" || at.Args[0].Text == "1") {
                                    if len(at.Args) > 1 {
                                        pol := at.Args[1].Text
                                        if pol == "dropOldest" || pol == "dropNewest" { tiny = true }
                                    }
                                }
                            }
                        }
                        if tiny {
                            fmt.Fprintf(f, "; edge_tiny_buffer %s->%s\n", e.From, e.To)
                        }
                    }
                }
            }
        }
    }
    return out, f.Close()
}
