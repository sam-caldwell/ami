package sem

import (
    "time"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeMultiPath validates MultiPath usage and merge.* attributes on Collect steps.
// - edge.MultiPath/MultiPath only valid on Collect → E_MP_ONLY_COLLECT
// - merge.* unknown → E_MERGE_ATTR_UNKNOWN
// - merge.* invalid arity → E_MERGE_ATTR_ARGS
// - merge.Sort without a field → W_MERGE_SORT_NO_FIELD
func AnalyzeMultiPath(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // allowed merge attributes and arity constraints (min,max; -1 means unbounded)
    type ar struct{ min, max int }
    merges := map[string]ar{
        "merge.Sort":       {1, 2},
        "merge.Stable":     {0, 0},
        "merge.Key":        {1, 1},
        "merge.Dedup":      {0, 1},
        "merge.Window":     {1, 1},
        "merge.Watermark":  {2, 2},
        "merge.Timeout":    {1, 1},
        "merge.Buffer":     {1, 2},
        "merge.PartitionBy":{1, 1},
    }
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        for _, s := range pd.Stmts {
            st, ok := s.(*ast.StepStmt)
            if !ok { continue }
            // detect multipath on non-Collect
            for _, at := range st.Attrs {
                if at.Name == "edge.MultiPath" || at.Name == "MultiPath" {
                    if st.Name != "Collect" {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MP_ONLY_COLLECT", Message: "edge.MultiPath only valid on Collect nodes", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                    }
                }
            }
            if st.Name != "Collect" { continue }
            // validate merge.* attributes and basic normalization
            seen := map[string]string{}
            keyField := ""
            partitionField := ""
            for _, at := range st.Attrs {
                if strings.HasPrefix(at.Name, "merge.") {
                    if rng, ok := merges[at.Name]; ok {
                        argc := len(at.Args)
                        if at.Name == "merge.Sort" && argc == 0 {
                            out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_SORT_NO_FIELD", Message: "merge.Sort requires a field", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                            continue
                        }
                        if argc < rng.min || (rng.max >= 0 && argc > rng.max) {
                            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_ARGS", Message: at.Name + ": invalid number of arguments", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                            continue
                        }
                        // additional validations
                        switch at.Name {
                        case "merge.Sort":
                            if argc >= 1 && strings.TrimSpace(at.Args[0].Text) == "" {
                                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_REQUIRED", Message: "merge.Sort: field is required", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                            }
                            if argc >= 2 {
                                ord := at.Args[1].Text
                                if ord != "asc" && ord != "desc" {
                                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_ARGS", Message: "merge.Sort: order must be 'asc' or 'desc'", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                                }
                            }
                        case "merge.Watermark":
                            if argc >= 1 && strings.TrimSpace(at.Args[0].Text) == "" {
                                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_REQUIRED", Message: "merge.Watermark: field is required", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                            }
                        case "merge.Window":
                            if argc >= 1 {
                                if at.Args[0].Text == "0" || strings.HasPrefix(at.Args[0].Text, "-") {
                                    out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_WINDOW_ZERO_OR_NEGATIVE", Message: "merge.Window: size should be > 0", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                                }
                            }
                        case "merge.Key", "merge.PartitionBy":
                            if argc >= 1 && strings.TrimSpace(at.Args[0].Text) == "" {
                                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_REQUIRED", Message: at.Name + ": field is required", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                            }
                        case "merge.Buffer":
                            if argc >= 1 {
                                if at.Args[0].Text == "0" || at.Args[0].Text == "1" {
                                    if argc >= 2 {
                                        pol := at.Args[1].Text
                                        if pol == "dropOldest" || pol == "dropNewest" {
                                            out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_TINY_BUFFER", Message: "merge.Buffer: tiny capacity with dropping policy", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                                        }
                                    }
                                }
                            }
                        }
                        // track fields for combo checks
                        if at.Name == "merge.Key" && argc >= 1 { keyField = at.Args[0].Text }
                        if at.Name == "merge.PartitionBy" && argc >= 1 { partitionField = at.Args[0].Text }
                        // conflict detection on repeated attributes with differing normalized value
                        key := at.Name
                        val := canonicalAttrValue(at.Name, at.Args)
                        if prev, ok := seen[key]; ok {
                            if prev != val {
                                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_CONFLICT", Message: at.Name + ": conflicting values", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                            }
                        } else {
                            seen[key] = val
                        }
                    } else {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_UNKNOWN", Message: "unknown merge attribute: " + at.Name, Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}})
                    }
                }
            }
            // cross-attribute conflict: PartitionBy vs Key with different fields (scaffold)
            if keyField != "" && partitionField != "" && keyField != partitionField {
                p := stepPos(st)
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_CONFLICT", Message: "merge.PartitionBy vs merge.Key conflict", Pos: &p})
            }
        }
    }
    return out
}

func canonicalAttrValue(name string, args []ast.Arg) string {
    // normalize value strings per attribute for conflict checks
    if name == "merge.Sort" {
        // field[/order]
        f := ""
        ord := ""
        if len(args) > 0 { f = args[0].Text }
        if len(args) > 1 { ord = args[1].Text }
        return f + "/" + ord
    }
    if name == "merge.Buffer" {
        cap := ""
        pol := ""
        if len(args) > 0 { cap = args[0].Text }
        if len(args) > 1 { pol = args[1].Text }
        return cap + "/" + pol
    }
    if len(args) > 0 { return args[0].Text }
    return ""
}

func stepPos(st *ast.StepStmt) diag.Position {
    return diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}
}
