package sem

import (
    "strconv"
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeEdgesInContext validates edge.* attributes using a package-level context of
// pipeline egress types. When egressType is nil, it falls back to scanning the file
// to build a local map (same as AnalyzeEdges).
func AnalyzeEdgesInContext(f *ast.File, egressType map[string]string) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()

    // If no context provided, build from this file (compat path)
    if egressType == nil {
        egressType = map[string]string{}
        for _, d := range f.Decls {
            if pd, ok := d.(*ast.PipelineDecl); ok {
                t := ""
                for _, s := range pd.Stmts {
                    if st, ok := s.(*ast.StepStmt); ok {
                        if strings.ToLower(st.Name) == "egress" {
                            for _, at := range st.Attrs {
                                if at.Name == "type" || at.Name == "Type" {
                                    if len(at.Args) > 0 { t = at.Args[0].Text }
                                }
                            }
                        }
                    }
                }
                egressType[pd.Name] = t
            }
        }
    }

    parseKV := func(args []ast.Arg) map[string]string {
        m := map[string]string{}
        for _, a := range args {
            s := a.Text
            if eq := strings.IndexByte(s, '='); eq > 0 {
                k := strings.TrimSpace(s[:eq])
                v := strings.TrimSpace(s[eq+1:])
                m[k] = v
            }
        }
        return m
    }

    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        for _, s := range pd.Stmts {
            st, ok := s.(*ast.StepStmt)
            if !ok { continue }
            for _, at := range st.Attrs {
                switch at.Name {
                case "edge.FIFO", "edge.LIFO":
                    kv := parseKV(at.Args)
                    for k := range kv { if k != "min" && k != "max" && k != "backpressure" {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EDGE_PARAM_UNKNOWN", Message: "unknown edge parameter: " + k, Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                    } }
                    if v, ok := kv["min"]; ok {
                        if n, err := strconv.Atoi(v); err != nil || n < 0 {
                            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EDGE_CAPACITY_INVALID", Message: "min must be non-negative", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                        }
                    }
                    if vmax, ok := kv["max"]; ok {
                        if nmax, err := strconv.Atoi(vmax); err != nil || nmax < 0 {
                            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EDGE_CAPACITY_INVALID", Message: "max must be non-negative integer", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                        } else if vmin, ok := kv["min"]; ok {
                            if nmin, err := strconv.Atoi(vmin); err == nil && nmax < nmin {
                                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EDGE_CAPACITY_ORDER", Message: "max must be >= min", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                            }
                        }
                    }
                    if bp, ok := kv["backpressure"]; ok {
                        switch bp {
                        case "block", "dropOldest", "dropNewest", "shuntNewest", "shuntOldest":
                        case "drop":
                            out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_EDGE_BP_LEGACY_DROP", Message: "legacy 'drop' alias; use dropOldest/dropNewest", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                        default:
                            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EDGE_BACKPRESSURE", Message: "invalid backpressure policy", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                        }
                    }
                case "edge.Pipeline":
                    kv := parseKV(at.Args)
                    name := kv["name"]
                    if name == "" {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EDGE_PIPE_NAME_REQUIRED", Message: "edge.Pipeline requires name=...", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                        continue
                    }
                    tgt, ok := egressType[name]
                    if !ok {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EDGE_PIPE_NOT_FOUND", Message: "edge.Pipeline target not found: " + name, Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                        continue
                    }
                    if want := kv["type"]; want != "" && tgt != "" {
                        if !typesCompatible(want, tgt) {
                            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_EDGE_PIPE_TYPE_MISMATCH", Message: "edge.Pipeline type mismatch", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"want": want, "target": tgt}})
                        }
                    }
                }
            }
        }
    }
    return out
}

