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
