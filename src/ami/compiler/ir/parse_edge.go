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

func atoiSafe(s string) int { n, _ := strconv.Atoi(s); return n }
