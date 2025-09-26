package codegen

import (
    "fmt"
    "sort"
    "strings"

    edg "github.com/sam-caldwell/ami/src/ami/compiler/edge"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// edgeInitPseudo renders a single edge initialization pseudo-instruction for
// listings. Concrete implementations are specialized per payload type.
func edgeInitPseudo(pipe string, idx int, s edg.Spec) string {
    switch v := s.(type) {
    case *edg.FIFO:
        return fmt.Sprintf("edge_init label=%s.step%d.in kind=fifo min=%d max=%d bp=%s type=%s",
            pipe, idx, v.MinCapacity, v.MaxCapacity, v.Backpressure, v.TypeName)
    case *edg.LIFO:
        return fmt.Sprintf("edge_init label=%s.step%d.in kind=lifo min=%d max=%d bp=%s type=%s",
            pipe, idx, v.MinCapacity, v.MaxCapacity, v.Backpressure, v.TypeName)
    case *edg.Pipeline:
        return fmt.Sprintf("edge_init label=%s.step%d.in kind=pipeline upstream=%s min=%d max=%d bp=%s type=%s",
            pipe, idx, v.UpstreamName, v.MinCapacity, v.MaxCapacity, v.Backpressure, v.TypeName)
    default:
        return fmt.Sprintf("edge_init label=%s.step%d.in kind=%s", pipe, idx, s.Kind())
    }
}

// writeMultiPath emits no-op pseudo-ops for MultiPath scaffolding to aid
// future integration and debugging. It does not affect runtime semantics.
func writeMultiPath(b *strings.Builder, pipe string, idx int, st ir.StepIR) {
    b.WriteString("  mp_begin label=")
    b.WriteString(fmt.Sprintf("%s.step%d.in", pipe, idx))
    b.WriteString("\n")
    // If a normalized config is present, emit a deterministic mp_cfg line
    if st.InMulti != nil {
        cfg := st.InMulti.Config
        if cfg == nil {
            // Derive a normalized config on-the-fly for listing purposes
            cfg = normalizeMergeOps(st.InMulti.Merge)
        }
        if cfg != nil {
        // collect key=value pairs deterministically
        kv := map[string]string{}
        if cfg.SortField != "" { kv["sort.field"] = cfg.SortField }
        if cfg.SortOrder != "" { kv["sort.order"] = cfg.SortOrder }
        if cfg.Stable { kv["stable"] = "true" }
        if cfg.Key != "" { kv["key.field"] = cfg.Key }
        if cfg.Dedup { kv["dedup"] = "true" }
        if cfg.DedupField != "" { kv["dedup.field"] = cfg.DedupField }
        if cfg.Window > 0 { kv["window"] = fmt.Sprintf("%d", cfg.Window) }
        if cfg.WatermarkField != "" { kv["watermark.field"] = cfg.WatermarkField }
        if cfg.WatermarkLateness != "" { kv["watermark.lateness"] = cfg.WatermarkLateness }
        if cfg.TimeoutMs > 0 { kv["timeout.ms"] = fmt.Sprintf("%d", cfg.TimeoutMs) }
        if cfg.BufferCapacity > 0 { kv["buffer.capacity"] = fmt.Sprintf("%d", cfg.BufferCapacity) }
        if cfg.BufferBackpressure != "" { kv["buffer.bp"] = cfg.BufferBackpressure }
        if cfg.PartitionBy != "" { kv["partitionBy.field"] = cfg.PartitionBy }
            if len(kv) > 0 {
            keys := make([]string, 0, len(kv))
            for k := range kv { keys = append(keys, k) }
            sort.Strings(keys)
            b.WriteString("  mp_cfg ")
            for i, k := range keys {
                if i > 0 { b.WriteByte(' ') }
                b.WriteString(k)
                b.WriteByte('=')
                b.WriteString(kv[k])
            }
            b.WriteString("\n")
            }
        }
    }
    for _, in := range st.InMulti.Inputs {
        b.WriteString("  mp_input ")
        b.WriteString(edgeInitPseudo(pipe, idx, in))
        b.WriteString("\n")
    }
    for _, op := range st.InMulti.Merge {
        b.WriteString("  mp_merge name=")
        b.WriteString(op.Name)
        if op.Raw != "" { b.WriteString(" args="); b.WriteString(op.Raw) }
        b.WriteString("\n")
    }
    b.WriteString("  mp_end label=")
    b.WriteString(fmt.Sprintf("%s.step%d.in", pipe, idx))
    b.WriteString("\n")
}

// normalizeMergeOps mirrors ir.normalizeMergeOps sufficiently for listing.
func normalizeMergeOps(ops []ir.MergeOpIR) *ir.MergeConfigIR {
    if len(ops) == 0 { return nil }
    cfg := &ir.MergeConfigIR{}
    split := func(s string) []string {
        return ir.SplitTopLevelCommasForCodegen(s)
    }
    trimq := func(s string) string {
        s = strings.TrimSpace(s)
        if len(s) >= 2 && ((s[0]=='"' && s[len(s)-1]=='"') || (s[0]=='\'' && s[len(s)-1]=='\'')) { return s[1:len(s)-1] }
        return s
    }
    for _, op := range ops {
        name := strings.ToLower(strings.TrimPrefix(op.Name, "merge."))
        args := split(op.Raw)
        switch name {
        case "sort":
            if len(args) >= 1 { cfg.SortField = trimq(args[0]) }
            if len(args) >= 2 { cfg.SortOrder = strings.ToLower(trimq(args[1])) }
            if cfg.SortOrder == "" { cfg.SortOrder = "asc" }
        case "stable":
            cfg.Stable = true
        case "key":
            if len(args) >= 1 { cfg.Key = trimq(args[0]) }
        case "dedup":
            cfg.Dedup = true
            if len(args) >= 1 { cfg.DedupField = trimq(args[0]) }
        case "window":
            if len(args) >= 1 { cfg.Window = atoiSafeForCodegen(trimq(args[0])) }
        case "watermark":
            if len(args) >= 1 { cfg.WatermarkField = trimq(args[0]) }
            if len(args) >= 2 { cfg.WatermarkLateness = trimq(args[1]) }
        case "timeout":
            if len(args) >= 1 { cfg.TimeoutMs = atoiSafeForCodegen(trimq(args[0])) }
        case "buffer":
            if len(args) >= 1 { cfg.BufferCapacity = atoiSafeForCodegen(trimq(args[0])) }
            if len(args) >= 2 { cfg.BufferBackpressure = trimq(args[1]) }
        case "partitionby":
            if len(args) >= 1 { cfg.PartitionBy = trimq(args[0]) }
        }
    }
    return cfg
}

func atoiSafeForCodegen(s string) int { var v int; _, _ = fmt.Sscanf(s, "%d", &v); return v }
