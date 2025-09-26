package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "strconv"
    "strings"
)

// analyzeEdges validates declarative edge specs provided via `in=edge.*(...)` args.
// Emits diagnostics for invalid capacity ordering, negative capacities, and unknown
// backpressure policies. For edge.Pipeline, also requires a non-empty name.
func analyzeEdges(pd astpkg.PipelineDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    // Validate MultiPath usage when present
    diags = append(diags, analyzeMultiPath(pd)...)
    parseFromNode := func(st astpkg.NodeCall) (interface{}, bool) {
        if v := strings.TrimSpace(st.Attrs["in"]); v != "" {
            if spec, ok := parseEdgeSpecFromValue(v); ok { return spec, true }
        }
        return parseEdgeSpecFromArgs(st.Args)
    }
    checkNode := func(st astpkg.NodeCall) {
        if spec, ok := parseFromNode(st); ok {
            switch v := spec.(type) {
            case fifoSpec:
                if v.Min < 0 {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_MINCAP", Message: "edge FIFO: minCapacity must be >= 0"})
                }
                if v.Max < v.Min {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_CAP_ORDER", Message: "edge FIFO: maxCapacity must be >= minCapacity"})
                }
                if v.BP != "" && v.BP != "block" && v.BP != "dropOldest" && v.BP != "dropNewest" {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_BP_INVALID", Message: "edge FIFO: invalid backpressure policy"})
                }
            case lifoSpec:
                if v.Min < 0 {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_MINCAP", Message: "edge LIFO: minCapacity must be >= 0"})
                }
                if v.Max < v.Min {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_CAP_ORDER", Message: "edge LIFO: maxCapacity must be >= minCapacity"})
                }
                if v.BP != "" && v.BP != "block" && v.BP != "dropOldest" && v.BP != "dropNewest" {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_BP_INVALID", Message: "edge LIFO: invalid backpressure policy"})
                }
            case pipeSpec:
                if v.Name == "" {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_NAME_REQUIRED", Message: "edge Pipeline: upstream name required"})
                }
                if v.Min < 0 {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_MINCAP", Message: "edge Pipeline: minCapacity must be >= 0"})
                }
                if v.Max < v.Min {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_CAP_ORDER", Message: "edge Pipeline: maxCapacity must be >= minCapacity"})
                }
                if v.BP != "" && v.BP != "block" && v.BP != "dropOldest" && v.BP != "dropNewest" {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_BP_INVALID", Message: "edge Pipeline: invalid backpressure policy"})
                }
            }
        }
    }
    for _, st := range pd.Steps { checkNode(st) }
    for _, st := range pd.ErrorSteps { checkNode(st) }
    return diags
}

// Minimal local spec structs to avoid cross-package dependency
type fifoSpec struct {
    Min, Max int
    BP, Type string
}
type lifoSpec struct {
    Min, Max int
    BP, Type string
}
type pipeSpec struct {
    Name     string
    Min, Max int
    BP, Type string
}

// parseEdgeSpecFromArgs: copy of tolerant parser used in IR lowering (simplified)
func parseEdgeSpecFromArgs(args []string) (interface{}, bool) {
    for _, a := range args {
        s := strings.TrimSpace(a)
        if !strings.HasPrefix(s, "in=") {
            continue
        }
        v := strings.TrimPrefix(s, "in=")
        if strings.HasPrefix(v, "edge.FIFO(") && strings.HasSuffix(v, ")") {
            params := parseKVList(v[len("edge.FIFO(") : len(v)-1])
            var f fifoSpec
            for k, val := range params {
                switch k {
                case "minCapacity":
                    f.Min = atoiSafe(val)
                case "maxCapacity":
                    f.Max = atoiSafe(val)
                case "backpressure":
                    f.BP = val
                case "type":
                    f.Type = val
                }
            }
            return f, true
        }
        if strings.HasPrefix(v, "edge.LIFO(") && strings.HasSuffix(v, ")") {
            params := parseKVList(v[len("edge.LIFO(") : len(v)-1])
            var l lifoSpec
            for k, val := range params {
                switch k {
                case "minCapacity":
                    l.Min = atoiSafe(val)
                case "maxCapacity":
                    l.Max = atoiSafe(val)
                case "backpressure":
                    l.BP = val
                case "type":
                    l.Type = val
                }
            }
            return l, true
        }
        if strings.HasPrefix(v, "edge.Pipeline(") && strings.HasSuffix(v, ")") {
            params := parseKVList(v[len("edge.Pipeline(") : len(v)-1])
            var p pipeSpec
            for k, val := range params {
                switch k {
                case "name":
                    p.Name = val
                case "minCapacity":
                    p.Min = atoiSafe(val)
                case "maxCapacity":
                    p.Max = atoiSafe(val)
                case "backpressure":
                    p.BP = val
                case "type":
                    p.Type = val
                }
            }
            return p, true
        }
    }
    return nil, false
}

// --- MultiPath tolerant parser and validations ---

type mpSpec struct {
    Inputs []interface{}
    Merge  []mergeOpSpec
}

type mergeOpSpec struct {
    Name string
    Raw  string
}

func parseMultiPathFromValue(v string) (mpSpec, bool) {
    s := strings.TrimSpace(v)
    if !strings.HasPrefix(s, "edge.MultiPath(") || !strings.HasSuffix(s, ")") { return mpSpec{}, false }
    inner := s[len("edge.MultiPath(") : len(s)-1]
    // locate inputs=[ ... ]
    idx := strings.Index(inner, "inputs=")
    if idx < 0 { return mpSpec{}, false }
    after := strings.TrimSpace(inner[idx+len("inputs="):])
    if len(after) == 0 || after[0] != '[' { return mpSpec{}, false }
    // capture bracket block
    i := 1
    depth := 1
    for i < len(after) && depth > 0 {
        switch after[i] {
        case '[': depth++
        case ']': depth--
        }
        i++
    }
    if depth != 0 { return mpSpec{}, false }
    list := after[1 : i-1]
    parts := splitTopLevelCommas(list)
    var inputs []interface{}
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p == "" { continue }
        if spec, ok := parseEdgeSpecFromValue(p); ok {
            inputs = append(inputs, spec)
        }
    }
    // one or more merge=Name(args) occurrences
    rest := strings.TrimSpace(after[i:])
    var merge []mergeOpSpec
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
            merge = append(merge, mergeOpSpec{Name: name, Raw: args})
            idx += m
        } else {
            break
        }
    }
    return mpSpec{Inputs: inputs, Merge: merge}, true
}

// analyzeMultiPath enforces minimal multi-path rules:
// - Only valid on Collect nodes.
// - inputs must be non-empty; inputs[0] must be a default upstream edge (FIFO).
// - All inputs must agree on payload Type when specified.
func analyzeMultiPath(pd astpkg.PipelineDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    // helpers
    splitArgs := func(raw string) []string { return splitTopLevelCommas(raw) }
    trimq := func(s string) string {
        s = strings.TrimSpace(s)
        if len(s) >= 2 && ((s[0]=='"' && s[len(s)-1]=='"') || (s[0]=='\'' && s[len(s)-1]=='\'')) { return s[1:len(s)-1] }
        return s
    }
    check := func(st astpkg.NodeCall) {
        v := strings.TrimSpace(st.Attrs["in"])
        if v == "" || !strings.HasPrefix(v, "edge.MultiPath(") { return }
        // Only on collect
        if strings.ToLower(st.Name) != "collect" {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MP_ONLY_COLLECT", Message: "edge.MultiPath is only valid on Collect nodes"})
        }
        mp, ok := parseMultiPathFromValue(v)
        if !ok {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MP_INVALID", Message: "invalid edge.MultiPath specification"})
            return
        }
        if len(mp.Inputs) == 0 {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MP_INPUTS_EMPTY", Message: "edge.MultiPath requires at least one input"})
            return
        }
        // First input must be FIFO default upstream edge
        switch mp.Inputs[0].(type) {
        case fifoSpec:
            // ok
        default:
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MP_INPUT0_KIND", Message: "first MultiPath input must be a FIFO default upstream edge"})
        }
        // Type compatibility across inputs (when provided)
        var base string
        for _, in := range mp.Inputs {
            var t string
            switch v := in.(type) {
            case fifoSpec:
                t = v.Type
            case lifoSpec:
                t = v.Type
            case pipeSpec:
                t = v.Type
            }
            if t == "" { continue }
            if base == "" {
                base = t
                continue
            }
            if base != t {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MP_INPUT_TYPE_MISMATCH", Message: "MultiPath inputs have mismatched payload types"})
                break
            }
        }
        // Merge op validation and smells
        allowed := map[string]bool{
            "sort": true, "stable": true, "key": true, "dedup": true, "window": true, "watermark": true, "timeout": true, "buffer": true, "partitionby": true,
        }
        // Track conflicts across attributes (last-write-wins allowed, but conflicting different values flagged)
        seen := map[string]string{}
        record := func(k, v string) {
            if v == "" { return }
            if prev, ok := seen[k]; ok && prev != v {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MERGE_ATTR_CONFLICT", Message: "conflicting merge attributes"})
            }
            seen[k] = v
        }
        for _, m := range mp.Merge {
            name := strings.ToLower(strings.TrimPrefix(m.Name, "merge."))
            if name == "" || !allowed[name] {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MERGE_ATTR_UNKNOWN", Message: "unknown merge attribute in MultiPath"})
                continue
            }
            args := splitArgs(m.Raw)
            switch name {
            case "sort":
                if len(args) == 0 || strings.TrimSpace(args[0]) == "" {
                    diags = append(diags, diag.Diagnostic{Level: diag.Warn, Code: "W_MERGE_SORT_NO_FIELD", Message: "merge.Sort requires a field"})
                } else if len(args) >= 2 {
                    ord := strings.ToLower(trimq(args[1]))
                    if ord != "asc" && ord != "desc" && ord != "" {
                        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MERGE_ATTR_ARGS", Message: "merge.Sort order must be asc or desc"})
                    }
                    record("sort.order", ord)
                }
                if len(args) >= 1 { record("sort.field", trimq(args[0])) }
            case "stable":
                if len(args) != 0 {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MERGE_ATTR_ARGS", Message: "merge.Stable takes no arguments"})
                }
                record("stable", "true")
            case "key":
                if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MERGE_ATTR_ARGS", Message: "merge.Key requires one field"})
                }
                if len(args) >= 1 { record("key", trimq(args[0])) }
            case "dedup":
                if len(args) > 1 {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MERGE_ATTR_ARGS", Message: "merge.Dedup accepts at most one optional field"})
                }
                record("dedup", "true")
                if len(args) >= 1 { record("dedup.field", trimq(args[0])) }
            case "window":
                if len(args) != 1 {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MERGE_ATTR_ARGS", Message: "merge.Window requires one integer size"})
                } else if atoiSafe(trimq(args[0])) <= 0 {
                    diags = append(diags, diag.Diagnostic{Level: diag.Warn, Code: "W_MERGE_WINDOW_ZERO_OR_NEGATIVE", Message: "merge.Window should be > 0"})
                }
                if len(args) == 1 { record("window", trimq(args[0])) }
            case "watermark":
                if len(args) < 1 {
                    diags = append(diags, diag.Diagnostic{Level: diag.Warn, Code: "W_MERGE_WATERMARK_MISSING_FIELD", Message: "merge.Watermark requires a field"})
                }
                if len(args) >= 1 { record("watermark.field", trimq(args[0])) }
                if len(args) >= 2 { record("watermark.lateness", trimq(args[1])) }
            case "timeout":
                if len(args) != 1 {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MERGE_ATTR_ARGS", Message: "merge.Timeout requires one integer ms"})
                }
                if len(args) == 1 { record("timeout", trimq(args[0])) }
            case "buffer":
                if len(args) < 1 || len(args) > 2 {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MERGE_ATTR_ARGS", Message: "merge.Buffer requires capacity and optional backpressure"})
                } else {
                    capv := atoiSafe(trimq(args[0]))
                    bp := ""
                    if len(args) == 2 { bp = strings.ToLower(trimq(args[1])) }
                    if (bp == "drop" || bp == "dropoldest" || bp == "dropnewest") && capv <= 1 && capv > 0 {
                        diags = append(diags, diag.Diagnostic{Level: diag.Warn, Code: "W_MERGE_TINY_BUFFER", Message: "merge.Buffer capacity<=1 with drop policy may cause loss"})
                    }
                    record("buffer.capacity", trimq(args[0]))
                    if bp != "" { record("buffer.bp", bp) }
                }
            case "partitionby":
                if len(args) != 1 || strings.TrimSpace(args[0]) == "" {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_MERGE_ATTR_ARGS", Message: "merge.PartitionBy requires one field"})
                }
                if len(args) == 1 { record("partitionBy", trimq(args[0])) }
            }
        }
    }
    for _, st := range pd.Steps { check(st) }
    for _, st := range pd.ErrorSteps { check(st) }
    return diags
}

// parseEdgeSpecFromValue parses an edge spec given the value part (e.g., "edge.FIFO(...)").
func parseEdgeSpecFromValue(v string) (interface{}, bool) {
    s := strings.TrimSpace(v)
    if strings.HasPrefix(s, "edge.FIFO(") && strings.HasSuffix(s, ")") {
        params := parseKVList(s[len("edge.FIFO(") : len(s)-1])
        var f fifoSpec
        for k, val := range params {
            switch k {
            case "minCapacity": f.Min = atoiSafe(val)
            case "maxCapacity": f.Max = atoiSafe(val)
            case "backpressure": f.BP = val
            case "type": f.Type = val
            }
        }
        return f, true
    }
    if strings.HasPrefix(s, "edge.LIFO(") && strings.HasSuffix(s, ")") {
        params := parseKVList(s[len("edge.LIFO(") : len(s)-1])
        var l lifoSpec
        for k, val := range params {
            switch k {
            case "minCapacity": l.Min = atoiSafe(val)
            case "maxCapacity": l.Max = atoiSafe(val)
            case "backpressure": l.BP = val
            case "type": l.Type = val
            }
        }
        return l, true
    }
    if strings.HasPrefix(s, "edge.Pipeline(") && strings.HasSuffix(s, ")") {
        params := parseKVList(s[len("edge.Pipeline(") : len(s)-1])
        var p pipeSpec
        for k, val := range params {
            switch k {
            case "name": p.Name = val
            case "minCapacity": p.Min = atoiSafe(val)
            case "maxCapacity": p.Max = atoiSafe(val)
            case "backpressure": p.BP = val
            case "type": p.Type = val
            }
        }
        return p, true
    }
    return nil, false
}

func parseKVList(s string) map[string]string {
    out := map[string]string{}
    parts := splitTopLevelCommas(s)
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p == "" {
            continue
        }
        if eq := strings.IndexByte(p, '='); eq >= 0 {
            k := strings.TrimSpace(p[:eq])
            v := strings.TrimSpace(p[eq+1:])
            // Trim optional quotes around value
            if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
                v = v[1 : len(v)-1]
            }
            out[k] = v
        }
    }
    return out
}

func splitTopLevelCommas(s string) []string {
    var parts []string
    depth := 0
    start := 0
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
                parts = append(parts, s[start:i])
                start = i + 1
            }
        }
    }
    // tail
    if start <= len(s) {
        parts = append(parts, s[start:])
    }
    return parts
}

func atoiSafe(s string) int { n, _ := strconv.Atoi(s); return n }

// typeRefToString renders a TypeRef to a string including pointer, slice, and generics.
func typeRefToString(t astpkg.TypeRef) string {
    var b strings.Builder
    if t.Ptr {
        b.WriteByte('*')
    }
    if t.Slice {
        b.WriteString("[]")
    }
    b.WriteString(t.Name)
    if len(t.Args) > 0 {
        b.WriteByte('<')
        for i, a := range t.Args {
            if i > 0 {
                b.WriteByte(',')
            }
            b.WriteString(typeRefToString(a))
        }
        b.WriteByte('>')
    }
    return b.String()
}
