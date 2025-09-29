package sem

import (
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeMergeFieldTypes performs conservative checks tying merge field usage
// to upstream Event<T> payload types when available via simple type("...") step
// attributes. It emits:
//   - E_MERGE_FIELD_ON_PRIMITIVE when upstream payload T is a primitive type
//     (bool,int,int64,float64,string) and a field is referenced.
//   - W_MERGE_FIELD_UNVERIFIED when no upstream type information is available.
// This is a scaffold until full payload typing is integrated.
func AnalyzeMergeFieldTypes(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()

    // Collect simple per-step declared types from type("...") attributes.
    stepType := map[string]string{}
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        for _, s := range pd.Stmts {
            if st, ok := s.(*ast.StepStmt); ok {
                for _, at := range st.Attrs {
                    if (at.Name == "type" || at.Name == "Type") && len(at.Args) > 0 && at.Args[0].Text != "" {
                        ts := at.Args[0].Text
                        if l := len(ts); l >= 2 {
                            if (ts[0] == '"' && ts[l-1] == '"') || (ts[0] == '\'' && ts[l-1] == '\'') { ts = ts[1:l-1] }
                        }
                        stepType[st.Name] = ts
                    }
                }
            }
        }
    }

    // helper to detect if Type is Event<primitive>
    isEventOfPrimitive := func(ts string) bool {
        ty, err := types.Parse(ts)
        if err != nil { return false }
        if g, ok := ty.(types.Generic); ok && g.Name == "Event" && len(g.Args) == 1 {
            switch a := g.Args[0].(type) {
            case types.Primitive:
                switch a.K {
                case types.Bool, types.Int, types.Int64, types.Float64, types.String:
                    return true
                }
            }
        }
        return false
    }

    // helper to resolve field path against payload types using types.Resolver
    fieldType := func(ts, field string) (types.Type, bool) {
        ty, err := types.Parse(ts)
        if err != nil { return nil, false }
        return types.ResolveField(ty, field)
    }

    // For each Collect, find upstreams and validate fields
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        // build list of edges targeting Collect
        var edgesToCollect []struct{ from string; pos diag.Position }
        for _, s := range pd.Stmts {
            if e, ok := s.(*ast.EdgeStmt); ok && e.To == "Collect" {
                edgesToCollect = append(edgesToCollect, struct{ from string; pos diag.Position }{from: e.From, pos: diag.Position{Line: e.Pos.Line, Column: e.Pos.Column, Offset: e.Pos.Offset}})
            }
        }
        if len(edgesToCollect) == 0 { continue }
        for _, s := range pd.Stmts {
            st, ok := s.(*ast.StepStmt)
            if !ok || st.Name != "Collect" { continue }
            // examine merge.Sort/Key/PartitionBy
            var fields []string
            for _, at := range st.Attrs {
                if (at.Name == "merge.Sort" || at.Name == "merge.Key" || at.Name == "merge.PartitionBy") && len(at.Args) > 0 {
                    fld := strings.TrimSpace(at.Args[0].Text)
                    if l := len(fld); l >= 2 {
                        if (fld[0] == '"' && fld[l-1] == '"') || (fld[0] == '\'' && fld[l-1] == '\'') { fld = fld[1:l-1] }
                    }
                    if fld != "" { fields = append(fields, fld) }
                }
            }
            if len(fields) == 0 { continue }
            // Determine if we have any upstream type info
            haveType := false
            primUp := false
            for _, e := range edgesToCollect {
                if ts := stepType[e.from]; ts != "" { haveType = true; if isEventOfPrimitive(ts) { primUp = true } }
            }
            if !haveType {
                p := diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}
                out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_FIELD_UNVERIFIED", Message: "merge field cannot be verified without upstream type information", Pos: &p})
                continue
            }
            if primUp {
                p := diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_FIELD_ON_PRIMITIVE", Message: "cannot reference field on Event<primitive> payload", Pos: &p})
                continue
            }
            // For each field, try resolve against any upstream with type info; require at least one success
            for _, fld := range fields {
                resolved := false
                orderable := false
                for _, e := range edgesToCollect {
                    if ts := stepType[e.from]; ts != "" {
                        if ft, ok := fieldType(ts, fld); ok {
                            resolved = true
                            orderable = types.IsOrderable(ft)
                            if orderable { break }
                        }
                    }
                }
                if !resolved {
                    p := diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_SORT_FIELD_UNKNOWN", Message: "merge field not found in payload", Pos: &p, Data: map[string]any{"field": fld}})
                    continue
                }
                if !orderable {
                    p := diag.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_SORT_FIELD_UNORDERABLE", Message: "merge field is not orderable", Pos: &p, Data: map[string]any{"field": fld}})
                }
            }
        }
    }
    return out
}
