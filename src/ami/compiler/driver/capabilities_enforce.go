package driver

import (
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// enforceCapabilitiesDriver performs a redundant, driver-stage capability/trust check
// to ensure enforcement even if semantic analyzers are bypassed in some flows.
// Rules:
//  - trust level 'untrusted' forbids io.* and net.* (E_TRUST_VIOLATION)
//  - io.* requires declared capability 'io' (E_CAPABILITY_REQUIRED)
//  - net.* requires declared capability 'net' (E_CAPABILITY_REQUIRED)
func enforceCapabilitiesDriver(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // Collect pragma caps/trust
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
    // Scan pipelines
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        for _, s := range pd.Stmts {
            st, ok := s.(*ast.StepStmt)
            if !ok { continue }
            name := strings.ToLower(st.Name)
            need := ""
            if strings.HasPrefix(name, "io.") { need = "io" }
            if strings.HasPrefix(name, "net.") { if need == "" { need = "net" } }
            if need == "" { continue }
            p := diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}
            if trust == "untrusted" {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_TRUST_VIOLATION", Message: "operation not allowed under trust level 'untrusted'", Pos: &p, Data: map[string]any{"node": st.Name, "required": need}})
                continue
            }
            if !caps[need] {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_CAPABILITY_REQUIRED", Message: "operation requires capability '" + need + "'", Pos: &p, Data: map[string]any{"node": st.Name, "required": need}})
            }
        }
    }
    return out
}

