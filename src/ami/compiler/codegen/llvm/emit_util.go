package llvm

import "github.com/sam-caldwell/ami/src/ami/compiler/ir"

// sanitizeIdent converts a string into a safe suffix for LLVM global names by
// replacing characters that are not valid in identifiers with underscores.
func sanitizeIdent(s string) string {
    if s == "" { return "_" }
    b := make([]byte, len(s))
    for i := 0; i < len(s); i++ {
        c := s[i]
        if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
            b[i] = c
        } else {
            b[i] = '_'
        }
    }
    return string(b)
}

// small helper to avoid importing strconv repeatedly in minimal emitter surface
func itoa(n int) string {
    if n == 0 { return "0" }
    var buf [20]byte
    i := len(buf)
    for n > 0 {
        i--
        buf[i] = byte('0' + (n % 10))
        n /= 10
    }
    return string(buf[i:])
}

// buildModuleMetaJSON assembles a compact JSON string summarizing module-level
// runtime-relevant metadata for downstream runtimes (scheduler, buffers, etc.).
func buildModuleMetaJSON(m ir.Module) string {
    // Minimal object builder (keep deterministic key order)
    b := make([]byte, 0, 512)
    appendStr := func(s string) { b = append(b, s...) }
    appendQuoted := func(s string) {
        b = append(b, '"')
        for i := 0; i < len(s); i++ { c := s[i]; if c == '"' || c == '\\' { b = append(b, '\\', c) } else { b = append(b, c) } }
        b = append(b, '"')
    }
    appendInt := func(n int) { appendStr(itoa(n)) }
    appendBool := func(v bool) { if v { appendStr("true") } else { appendStr("false") } }
    appendStr("{")
    appendQuoted("schema"); appendStr(":"); appendQuoted("ami.meta.v1")
    appendStr(","); appendQuoted("package"); appendStr(":"); appendQuoted(m.Package)
    if m.Concurrency > 0 { appendStr(","); appendQuoted("concurrency"); appendStr(":"); appendInt(m.Concurrency) }
    if m.Backpressure != "" { appendStr(","); appendQuoted("backpressure"); appendStr(":"); appendQuoted(m.Backpressure) }
    if m.Schedule != "" { appendStr(","); appendQuoted("schedule"); appendStr(":"); appendQuoted(m.Schedule) }
    if m.TelemetryEnabled { appendStr(","); appendQuoted("telemetryEnabled"); appendStr(":"); appendBool(true) }
    if l := len(m.Capabilities); l > 0 {
        appendStr(","); appendQuoted("capabilities"); appendStr(":[")
        for i, c := range m.Capabilities { if i > 0 { appendStr(",") }; appendQuoted(c) }
        appendStr("]")
    }
    if m.TrustLevel != "" { appendStr(","); appendQuoted("trustLevel"); appendStr(":"); appendQuoted(m.TrustLevel) }
    if l := len(m.Pipelines); l > 0 {
        appendStr(","); appendQuoted("pipelines"); appendStr(":[")
        for i, p := range m.Pipelines {
            if i > 0 { appendStr(",") }
            appendStr("{"); appendQuoted("name"); appendStr(":"); appendQuoted(p.Name)
            if len(p.Collect) > 0 {
                appendStr(","); appendQuoted("collect"); appendStr(":[")
                for j, cs := range p.Collect {
                    if j > 0 { appendStr(",") }
                    appendStr("{"); appendQuoted("step"); appendStr(":"); appendQuoted(cs.Step)
                    if cs.Merge != nil {
                        mp := cs.Merge
                        appendStr(","); appendQuoted("merge"); appendStr(":{")
                        // Buffer
                        if mp.Buffer.Capacity > 0 || mp.Buffer.Policy != "" {
                            appendQuoted("buffer"); appendStr(":{"); appendQuoted("capacity"); appendStr(":"); appendInt(mp.Buffer.Capacity)
                            if mp.Buffer.Policy != "" { appendStr(","); appendQuoted("policy"); appendStr(":"); appendQuoted(mp.Buffer.Policy) }
                            appendStr("}")
                            appendStr(",")
                        }
                        // Sort
                        if len(mp.Sort) > 0 {
                            appendQuoted("sort"); appendStr(":[")
                            for k, sk := range mp.Sort { if k > 0 { appendStr(",") }; appendStr("{"); appendQuoted("field"); appendStr(":"); appendQuoted(sk.Field); appendStr(","); appendQuoted("order"); appendStr(":"); appendQuoted(sk.Order); appendStr("}") }
                            appendStr("]"); appendStr(",")
                        }
                        // Stable
                        if mp.Stable { appendQuoted("stable"); appendStr(":true,") }
                        // Window/Watermark/Timeout
                        if mp.Window > 0 { appendQuoted("window"); appendStr(":"); appendInt(mp.Window); appendStr(",") }
                        if mp.Watermark != nil { appendQuoted("watermark"); appendStr(":{"); appendQuoted("field"); appendStr(":"); appendQuoted(mp.Watermark.Field); if mp.Watermark.LatenessMs > 0 { appendStr(","); appendQuoted("latenessMs"); appendStr(":"); appendInt(mp.Watermark.LatenessMs) }; appendStr("}"); appendStr(",") }
                        if mp.TimeoutMs > 0 { appendQuoted("timeoutMs"); appendStr(":"); appendInt(mp.TimeoutMs); appendStr(",") }
                        // PartitionBy/Key/Dedup/LatePolicy
                        if mp.PartitionBy != "" { appendQuoted("partitionBy"); appendStr(":"); appendQuoted(mp.PartitionBy); appendStr(",") }
                        if mp.Key != "" { appendQuoted("key"); appendStr(":"); appendQuoted(mp.Key); appendStr(",") }
                        if mp.DedupField != "" { appendQuoted("dedupField"); appendStr(":"); appendQuoted(mp.DedupField); appendStr(",") }
                        if mp.LatePolicy != "" { appendQuoted("latePolicy"); appendStr(":"); appendQuoted(mp.LatePolicy); appendStr(",") }
                        // trim trailing comma if present
                        if len(b) > 0 && b[len(b)-1] == ',' { b = b[:len(b)-1] }
                        appendStr("}")
                    }
                    appendStr("}")
                }
                appendStr("]")
            }
            appendStr("}")
        }
        appendStr("]")
    }
    if l := len(m.ErrorPipes); l > 0 {
        appendStr(","); appendQuoted("errorPipelines"); appendStr(":[")
        for i, ep := range m.ErrorPipes {
            if i > 0 { appendStr(",") }
            appendStr("{"); appendQuoted("pipeline"); appendStr(":"); appendQuoted(ep.Pipeline); appendStr(","); appendQuoted("steps"); appendStr(":[")
            for j, s := range ep.Steps { if j > 0 { appendStr(",") }; appendQuoted(s) }
            appendStr("]}")
        }
        appendStr("]")
    }
    appendStr("}")
    return string(b)
}
