package sem

import (
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeErrorSemantics validates error pipeline blocks inside pipeline declarations.
// - cannot start with ingress (E_ERRPIPE_START_INVALID)
// - must end with egress (E_ERRPIPE_END_EGRESS)
// - unknown nodes flagged as E_UNKNOWN_NODE
func AnalyzeErrorSemantics(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok || pd.Error == nil || pd.Error.Body == nil { continue }
        var steps []*ast.StepStmt
        for _, s := range pd.Error.Body.Stmts {
            if st, ok := s.(*ast.StepStmt); ok { steps = append(steps, st) }
        }
        if len(steps) == 0 { continue }
        // start invalid when ingress
        if strings.ToLower(steps[0].Name) == "ingress" {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_ERRPIPE_START_INVALID", Message: "error pipeline cannot start with 'ingress'", Pos: &diag.Position{Line: steps[0].Pos.Line, Column: steps[0].Pos.Column, Offset: steps[0].Pos.Offset}})
        }
        // end must be egress
        last := steps[len(steps)-1]
        if strings.ToLower(last.Name) != "egress" {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_ERRPIPE_END_EGRESS", Message: "error pipeline must end with 'egress'", Pos: &diag.Position{Line: last.Pos.Line, Column: last.Pos.Column, Offset: last.Pos.Offset}})
        }
        // unknown nodes
        for _, st := range steps {
            nameLower := strings.ToLower(st.Name)
            if nameLower != "ingress" && nameLower != "egress" && nameLower != "transform" && nameLower != "fanout" && nameLower != "collect" && nameLower != "mutable" {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_UNKNOWN_NODE", Message: "unknown node: " + st.Name, Pos: &diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}})
            }
        }
    }
    return out
}

