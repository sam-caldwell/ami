package driver

import (
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// analyzeCapabilityIR approximates an IR/codegen-stage capability check:
// flags io.* usage on pipeline steps outside ingress/egress positions.
// This complements the semantics-layer check and helps keep detection closer to lowering.
func analyzeCapabilityIR(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        // collect step stmts
        var steps []*ast.StepStmt
        for _, s := range pd.Stmts {
            if st, ok := s.(*ast.StepStmt); ok { steps = append(steps, st) }
        }
        if len(steps) == 0 { continue }
        for i, st := range steps {
            if strings.HasPrefix(st.Name, "io.") {
                // allowed only when first (ingress) or last (egress)
                if i != 0 && i != len(steps)-1 {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_IO_PERMISSION_IR", Message: "io.* operations only allowed in ingress/egress nodes (IR)", Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}})
                }
            }
        }
    }
    return out
}

