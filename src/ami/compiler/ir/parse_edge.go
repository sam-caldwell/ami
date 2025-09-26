package ir

import (
    "strconv"
    "strings"

    edg "github.com/sam-caldwell/ami/src/ami/compiler/edge"
)

// parseEdgeSpecFromArgs scans a node's raw arg list and extracts an edge.* spec
// from an `in=` parameter when present.
func parseEdgeSpecFromArgs(args []string) (edg.Spec, bool) {
    for _, a := range args {
        s := strings.TrimSpace(a)
        if !strings.HasPrefix(s, "in=") {
            continue
        }
        v := strings.TrimPrefix(s, "in=")
        if strings.HasPrefix(v, "edge.FIFO(") {
            params := parseKVList(v[len("edge.FIFO(") : len(v)-1])
            var f edg.FIFO
            for k, val := range params {
                switch k {
                case "minCapacity":
                    f.MinCapacity = atoiSafe(val)
                case "maxCapacity":
                    f.MaxCapacity = atoiSafe(val)
                case "backpressure":
                    f.Backpressure = edg.BackpressurePolicy(val)
                case "type":
                    f.TypeName = val
                }
            }
            return &f, true
        }
        if strings.HasPrefix(v, "edge.LIFO(") {
            params := parseKVList(v[len("edge.LIFO(") : len(v)-1])
            var l edg.LIFO
            for k, val := range params {
                switch k {
                case "minCapacity":
                    l.MinCapacity = atoiSafe(val)
                case "maxCapacity":
                    l.MaxCapacity = atoiSafe(val)
                case "backpressure":
                    l.Backpressure = edg.BackpressurePolicy(val)
                case "type":
                    l.TypeName = val
                }
            }
            return &l, true
        }
        if strings.HasPrefix(v, "edge.Pipeline(") {
            params := parseKVList(v[len("edge.Pipeline(") : len(v)-1])
            var p edg.Pipeline
            for k, val := range params {
                switch k {
                case "name":
                    p.UpstreamName = val
                case "minCapacity":
                    p.MinCapacity = atoiSafe(val)
                case "maxCapacity":
                    p.MaxCapacity = atoiSafe(val)
                case "backpressure":
                    p.Backpressure = edg.BackpressurePolicy(val)
                case "type":
                    p.TypeName = val
                }
            }
            return &p, true
        }
    }
    return nil, false
}

// parseEdgeSpecFromValue parses an edge spec given the value part (e.g., "edge.FIFO(...)").
func parseEdgeSpecFromValue(v string) (edg.Spec, bool) {
    s := strings.TrimSpace(v)
    if strings.HasPrefix(s, "edge.FIFO(") {
        params := parseKVList(s[len("edge.FIFO(") : len(s)-1])
        var f edg.FIFO
        for k, val := range params {
            switch k {
            case "minCapacity": f.MinCapacity = atoiSafe(val)
            case "maxCapacity": f.MaxCapacity = atoiSafe(val)
            case "backpressure": f.Backpressure = edg.BackpressurePolicy(val)
            case "type": f.TypeName = val
            }
        }
        return &f, true
    }
    if strings.HasPrefix(s, "edge.LIFO(") {
        params := parseKVList(s[len("edge.LIFO(") : len(s)-1])
        var l edg.LIFO
        for k, val := range params {
            switch k {
            case "minCapacity": l.MinCapacity = atoiSafe(val)
            case "maxCapacity": l.MaxCapacity = atoiSafe(val)
            case "backpressure": l.Backpressure = edg.BackpressurePolicy(val)
            case "type": l.TypeName = val
            }
        }
        return &l, true
    }
    if strings.HasPrefix(s, "edge.Pipeline(") {
        params := parseKVList(s[len("edge.Pipeline(") : len(s)-1])
        var p edg.Pipeline
        for k, val := range params {
            switch k {
            case "name": p.UpstreamName = val
            case "minCapacity": p.MinCapacity = atoiSafe(val)
            case "maxCapacity": p.MaxCapacity = atoiSafe(val)
            case "backpressure": p.Backpressure = edg.BackpressurePolicy(val)
            case "type": p.TypeName = val
            }
        }
        return &p, true
    }
    return nil, false
}

// --- MultiPath (scaffold) ---

// MultiPathIR is an IR-only scaffold for edge.MultiPath used in pipelines debug schema.
type MultiPathIR struct {
    Inputs []edg.Spec
    Merge  []MergeOpIR
    Config *MergeConfigIR
}

// MergeOpIR carries a merge operation name and raw argument string.
type MergeOpIR struct {
    Name string
    Raw  string
}

// parseMultiPathSpec parses a value starting with edge.MultiPath( ... ).
// Supported shape (tolerant): edge.MultiPath(inputs=[ <edgeSpec>{, <edgeSpec>} ], merge=Name(args...))
func parseMultiPathSpec(v string) (*MultiPathIR, bool) {
    s := strings.TrimSpace(v)
    if !strings.HasPrefix(s, "edge.MultiPath(") || !strings.HasSuffix(s, ")") { return nil, false }
    inner := s[len("edge.MultiPath(") : len(s)-1]
    // find inputs=[ ... ]
    idx := strings.Index(inner, "inputs=")
    if idx < 0 { return nil, false }
    after := inner[idx+len("inputs="):]
    after = strings.TrimSpace(after)
    if len(after) == 0 || after[0] != '[' { return nil, false }
    // capture bracketed list
    i := 1
    depth := 1
    for i < len(after) && depth > 0 {
        switch after[i] {
        case '[': depth++
        case ']': depth--
        }
        i++
    }
    if depth != 0 { return nil, false }
    list := after[1 : i-1]
    // split at top-level commas (respecting parentheses)
    parts := splitTopLevelCommas(list)
    var inputs []edg.Spec
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p == "" { continue }
        if spec, ok := parseEdgeSpecFromValue(p); ok {
            inputs = append(inputs, spec)
        }
    }
    mp := &MultiPathIR{Inputs: inputs}
    // Optional one or more merge=Name(args) entries in the remaining string
    rest := strings.TrimSpace(after[i:])
    // scan rest for top-level occurrences of "merge=" and capture Name(args)
    for idx := 0; idx < len(rest); {
        j := strings.Index(rest[idx:], "merge=")
        if j < 0 { break }
        idx += j + len("merge=")
        mv := strings.TrimSpace(rest[idx:])
        if k := strings.IndexByte(mv, '('); k > 0 {
            name := strings.TrimSpace(mv[:k])
            m := k + 1
            d := 1
            for m < len(mv) && d > 0 {
                if mv[m] == '(' { d++ }
                if mv[m] == ')' { d-- }
                m++
            }
            args := ""
            if d == 0 { args = mv[k+1 : m-1] }
            mp.Merge = append(mp.Merge, MergeOpIR{Name: name, Raw: args})
            idx += m
        } else {
            // malformed; stop
            break
        }
    }
    // Build normalized config (tolerant)
    if cfg := normalizeMergeOps(mp.Merge); cfg != nil {
        mp.Config = cfg
    }
    return mp, true
}

// parseKVList parses a simple comma-separated list of `key=value` entries.
func parseKVList(s string) map[string]string {
    out := map[string]string{}
    parts := splitTopLevelCommas(s)
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p == "" {
            continue
        }
        eq := strings.IndexByte(p, '=')
        if eq < 0 {
            continue
        }
        k := strings.TrimSpace(p[:eq])
        v := strings.TrimSpace(p[eq+1:])
        if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
            v = v[1 : len(v)-1]
        }
        out[k] = v
    }
    return out
}

func splitTopLevelCommas(s string) []string {
    var out []string
    depth := 0
    last := 0
    for i := 0; i < len(s); i++ {
        switch s[i] {
        case '(':
            depth++
        case ')':
            if depth > 0 {
                depth--
            }
        case ',':
            if depth == 0 {
                out = append(out, s[last:i])
                last = i + 1
            }
        }
    }
    out = append(out, s[last:])
    return out
}

// SplitTopLevelCommasForCodegen exposes the comma splitter for codegen normalization without creating a new dependency.
func SplitTopLevelCommasForCodegen(s string) []string { return splitTopLevelCommas(s) }

func atoiSafe(s string) int { n, _ := strconv.Atoi(s); return n }

// --- Merge normalization ---

// MergeConfigIR mirrors schemas.MergeConfigV1 but stays internal to IR.
type MergeConfigIR struct {
    SortField, SortOrder string
    Stable               bool
    Key                  string
    Dedup                bool
    DedupField           string
    Window               int
    WatermarkField       string
    WatermarkLateness    string
    TimeoutMs            int
    BufferCapacity       int
    BufferBackpressure   string
    PartitionBy          string
}

func normalizeMergeOps(ops []MergeOpIR) *MergeConfigIR {
    if len(ops) == 0 { return nil }
    cfg := &MergeConfigIR{}
    for _, op := range ops {
        name := strings.ToLower(strings.TrimPrefix(op.Name, "merge."))
        args := splitTopLevelCommas(op.Raw)
        trimq := func(s string) string {
            s = strings.TrimSpace(s)
            if len(s) >= 2 && ((s[0]=='"' && s[len(s)-1]=='"') || (s[0]=='\'' && s[len(s)-1]=='\'')) { return s[1:len(s)-1] }
            return s
        }
        switch name {
        case "sort":
            if len(args) >= 1 {
                cfg.SortField = trimq(args[0])
            }
            if len(args) >= 2 {
                cfg.SortOrder = strings.ToLower(trimq(args[1]))
            }
            if cfg.SortOrder == "" { cfg.SortOrder = "asc" }
        case "stable":
            cfg.Stable = true
        case "key":
            if len(args) >= 1 { cfg.Key = trimq(args[0]) }
        case "dedup":
            cfg.Dedup = true
            if len(args) >= 1 { cfg.DedupField = trimq(args[0]) }
        case "window":
            if len(args) >= 1 { cfg.Window = atoiSafe(trimq(args[0])) }
        case "watermark":
            if len(args) >= 1 { cfg.WatermarkField = trimq(args[0]) }
            if len(args) >= 2 { cfg.WatermarkLateness = trimq(args[1]) }
        case "timeout":
            if len(args) >= 1 { cfg.TimeoutMs = atoiSafe(trimq(args[0])) }
        case "buffer":
            if len(args) >= 1 { cfg.BufferCapacity = atoiSafe(trimq(args[0])) }
            if len(args) >= 2 { cfg.BufferBackpressure = trimq(args[1]) }
        case "partitionby":
            if len(args) >= 1 { cfg.PartitionBy = trimq(args[0]) }
        }
    }
    return cfg
}
