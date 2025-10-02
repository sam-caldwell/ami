package driver

import (
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

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
                if len(at.Args) >= 2 { if ms, ok := parseDurationMs(at.Args[1].Text); ok { wm.LatenessMs = ms } }
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

