package sem

import (
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeCapabilities enforces capability/trust constraints declared via pragmas.
// Pragmas:
//   #pragma capabilities list=io,net   (or args: capabilities io net)
//   #pragma trust level=trusted|untrusted
// Rules (scaffold, docx-aligned subset):
//   - io.* steps require declared capability 'io' (E_CAPABILITY_REQUIRED)
//   - trust level 'untrusted' forbids io.* anywhere (E_TRUST_VIOLATION)
// Node-position constraints for io.* are handled in AnalyzePipelineSemantics.
func AnalyzeCapabilities(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()

    // Collect pragmas
    caps := map[string]bool{}
    trust := ""
    for _, pr := range f.Pragmas {
        switch strings.ToLower(pr.Domain) {
        case "capabilities":
            if pr.Params != nil {
                if lst, ok := pr.Params["list"]; ok && lst != "" {
                    for _, t := range strings.Split(lst, ",") {
                        t = strings.TrimSpace(strings.ToLower(t))
                        if t != "" { caps[t] = true }
                    }
                }
            }
            for _, a := range pr.Args {
                t := strings.TrimSpace(strings.ToLower(a))
                if t != "" { caps[t] = true }
            }
        case "trust":
            if pr.Params != nil {
                if lv, ok := pr.Params["level"]; ok { trust = strings.ToLower(strings.TrimSpace(lv)) }
            }
        }
    }

    requiresCap := func(stepName string) string {
        ln := strings.ToLower(stepName)
        if strings.HasPrefix(ln, "io.") { return "io" }
        if strings.HasPrefix(ln, "net.") { return "net" }
        return ""
    }

    // Walk pipelines and enforce caps/trust
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        for _, s := range pd.Stmts {
            st, ok := s.(*ast.StepStmt)
            if !ok { continue }
            need := requiresCap(st.Name)
            if need == "" { continue }
            // Trust: untrusted forbids io/net
            if trust == "untrusted" {
                p := diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TRUST_VIOLATION", Message: "operation not allowed under trust level 'untrusted'", Pos: &p, Data: map[string]any{"node": st.Name, "required": need}})
                continue
            }
            if !caps[need] {
                p := diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CAPABILITY_REQUIRED", Message: "operation requires capability '" + need + "'", Pos: &p, Data: map[string]any{"node": st.Name, "required": need}})
            }
        }
    }
    return out
}

