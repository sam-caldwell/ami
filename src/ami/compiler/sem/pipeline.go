package sem

import (
    "time"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzePipelineSemantics enforces basic pipeline invariants:
// - first step must be `ingress` (E_PIPELINE_START_INGRESS)
// - last step must be `egress` (E_PIPELINE_END_EGRESS)
// - unknown steps emit E_UNKNOWN_NODE (only `ingress` and `egress` are recognized)
func AnalyzePipelineSemantics(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // scan for pipeline decls
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        // collect step statements only
        var steps []*ast.StepStmt
        for _, s := range pd.Stmts {
            if st, ok := s.(*ast.StepStmt); ok {
                steps = append(steps, st)
            }
        }
        // start check
        if len(steps) == 0 {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PIPELINE_START_INGRESS", Message: "pipeline must start with 'ingress'", File: "", Pos: &diag.Position{Line: pd.Pos.Line, Column: pd.Pos.Column, Offset: pd.Pos.Offset}})
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PIPELINE_END_EGRESS", Message: "pipeline must end with 'egress'", File: "", Pos: &diag.Position{Line: pd.Pos.Line, Column: pd.Pos.Column, Offset: pd.Pos.Offset}})
            continue
        }
        if steps[0].Name != "ingress" {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PIPELINE_START_INGRESS", Message: "pipeline must start with 'ingress'", File: "", Pos: &diag.Position{Line: steps[0].Pos.Line, Column: steps[0].Pos.Column, Offset: steps[0].Pos.Offset}})
        }
        // end check
        last := steps[len(steps)-1]
        if last.Name != "egress" {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PIPELINE_END_EGRESS", Message: "pipeline must end with 'egress'", File: "", Pos: &diag.Position{Line: last.Pos.Line, Column: last.Pos.Column, Offset: last.Pos.Offset}})
        }
        // position/uniqueness checks, io permissions, and unknown nodes
        ingressCount := 0
        egressCount := 0
        for _, st := range steps {
            if st.Name == "ingress" {
                if ingressCount > 0 { // duplicate beyond the first seen
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DUP_INGRESS", Message: "duplicate ingress", File: "", Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}})
                }
                // position must be 0
                // check based on index by scanning again to find this st; simpler: compare to steps[0]
                if st != steps[0] {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_INGRESS_POSITION", Message: "'ingress' only allowed at position 0", File: "", Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}})
                }
                ingressCount++
                continue
            }
            if st.Name == "egress" {
                if egressCount > 0 { // duplicate beyond the first seen
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_DUP_EGRESS", Message: "duplicate egress", File: "", Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}})
                }
                // position must be last
                if st != steps[len(steps)-1] {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EGRESS_POSITION", Message: "'egress' only allowed at last position", File: "", Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}})
                }
                egressCount++
                continue
            }
            // IO capability detection: any non-ingress/egress node using io.*
            if strings.HasPrefix(st.Name, "io.") {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_IO_PERMISSION", Message: "io.* operations only allowed in ingress/egress nodes", File: "", Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}})
            }
            // unknown nodes
            if st.Name != "ingress" && st.Name != "egress" {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_UNKNOWN_NODE", Message: "unknown node: " + st.Name, File: "", Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}})
            }
        }
    }
    return out
}
