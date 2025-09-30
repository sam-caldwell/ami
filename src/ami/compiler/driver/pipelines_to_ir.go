package driver

import (
    "strconv"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerPipelines extracts Collect/merge attributes into IR pipelines for a single AST file.
func lowerPipelines(f *ast.File) []ir.Pipeline {
    var out []ir.Pipeline
    if f == nil { return out }
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        pl := ir.Pipeline{Name: pd.Name}
        for _, s := range pd.Stmts {
            st, ok := s.(*ast.StepStmt)
            if !ok || st.Name != "Collect" { continue }
            if mp := toMergePlan(st); mp != nil {
                pl.Collect = append(pl.Collect, ir.CollectSpec{Step: st.Name, Merge: mp})
            }
        }
        if len(pl.Collect) > 0 { out = append(out, pl) }
    }
    return out
}

func toMergePlan(st *ast.StepStmt) *ir.MergePlan {
    if st == nil { return nil }
    mp := &ir.MergePlan{}
    saw := false
    for _, at := range st.Attrs {
        name := at.Name
        switch name {
        case "merge.Sort":
            if len(at.Args) >= 1 {
                f := trimQuotes(at.Args[0].Text)
                ord := "asc"
                if len(at.Args) >= 2 {
                    o := strings.ToLower(trimQuotes(at.Args[1].Text))
                    if o == "asc" || o == "desc" { ord = o }
                }
                mp.Sort = append(mp.Sort, ir.SortKey{Field: f, Order: ord})
                saw = true
            }
        case "merge.Stable":
            mp.Stable = true; saw = true
        case "merge.Key":
            if len(at.Args) >= 1 { mp.Key = trimQuotes(at.Args[0].Text); saw = true }
        case "merge.PartitionBy":
            if len(at.Args) >= 1 { mp.PartitionBy = trimQuotes(at.Args[0].Text); saw = true }
        case "merge.Dedup":
            if len(at.Args) >= 1 { mp.DedupField = trimQuotes(at.Args[0].Text) } else { mp.DedupField = "" }
            saw = true
        case "merge.Window":
            if len(at.Args) >= 1 {
                if n, ok := atoiSafe(at.Args[0].Text); ok && n > 0 { mp.Window = n; saw = true }
            }
        case "merge.Timeout":
            if len(at.Args) >= 1 {
                if n, ok := atoiSafe(at.Args[0].Text); ok && n > 0 { mp.TimeoutMs = n; saw = true }
            }
        case "merge.Watermark":
            if len(at.Args) >= 1 {
                fld := trimQuotes(at.Args[0].Text)
                wm := &ir.Watermark{Field: fld}
                if len(at.Args) >= 2 {
                    if ms, ok := parseDurationMs(at.Args[1].Text); ok { wm.LatenessMs = ms }
                }
                mp.Watermark = wm; saw = true
            }
        case "merge.Buffer":
            if len(at.Args) >= 1 { if n, ok := atoiSafe(at.Args[0].Text); ok { mp.Buffer.Capacity = n } }
            if len(at.Args) >= 2 { pol := strings.ToLower(trimQuotes(at.Args[1].Text)); mp.Buffer.Policy = pol }
            saw = true
        }
    }
    if !saw { return nil }
    return mp
}

func trimQuotes(s string) string {
    if l := len(s); l >= 2 {
        if (s[0] == '"' && s[l-1] == '"') || (s[0] == '\'' && s[l-1] == '\'') { return s[1:l-1] }
    }
    return s
}

func atoiSafe(s string) (int, bool) {
    s = strings.TrimSpace(trimQuotes(s))
    n, err := strconv.Atoi(s)
    if err != nil { return 0, false }
    return n, true
}

func parseDurationMs(s string) (int, bool) {
    s = strings.TrimSpace(trimQuotes(s))
    // Accept digits (ms) or digits with suffix ms/s/m/h
    if n, err := strconv.Atoi(s); err == nil { return n, true }
    // suffix
    mul := 1
    unit := "ms"
    if strings.HasSuffix(s, "ms") { unit = "ms"; mul = 1; s = strings.TrimSuffix(s, "ms") } else
    if strings.HasSuffix(s, "s") { unit = "s"; mul = 1000; s = strings.TrimSuffix(s, "s") } else
    if strings.HasSuffix(s, "m") { unit = "m"; mul = 60*1000; s = strings.TrimSuffix(s, "m") } else
    if strings.HasSuffix(s, "h") { unit = "h"; mul = 60*60*1000; s = strings.TrimSuffix(s, "h") }
    _ = unit
    n, err := strconv.Atoi(s)
    if err != nil { return 0, false }
    return n * mul, true
}

