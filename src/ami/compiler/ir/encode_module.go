package ir

import "encoding/json"

// EncodeModule produces deterministic JSON for debug.
func EncodeModule(m Module) ([]byte, error) {
    jm := map[string]any{
        "schema":   "ir.v1",
        "package":  m.Package,
        "functions": []any{},
    }
    if m.Concurrency > 0 { jm["concurrency"] = m.Concurrency }
    if m.Backpressure != "" { jm["backpressurePolicy"] = m.Backpressure }
    if m.TelemetryEnabled { jm["telemetryEnabled"] = true }
    if m.Schedule != "" { jm["schedule"] = m.Schedule }
    if len(m.Capabilities) > 0 { jm["capabilities"] = m.Capabilities }
    if m.TrustLevel != "" { jm["trustLevel"] = m.TrustLevel }
    if m.ExecContext != nil { jm["execContext"] = m.ExecContext }
    if m.EventMeta != nil { jm["eventmeta"] = map[string]any{"schema": m.EventMeta.Schema, "fields": m.EventMeta.Fields} }
    if len(m.Directives) > 0 {
        ds := make([]any, 0, len(m.Directives))
        for _, d := range m.Directives {
            obj := map[string]any{"domain": d.Domain}
            if d.Key != "" { obj["key"] = d.Key }
            if d.Value != "" { obj["value"] = d.Value }
            if len(d.Args) > 0 { obj["args"] = d.Args }
            if len(d.Params) > 0 { obj["params"] = d.Params }
            ds = append(ds, obj)
        }
        jm["directives"] = ds
    }
    if len(m.Pipelines) > 0 {
        ps := make([]any, 0, len(m.Pipelines))
        for _, p := range m.Pipelines {
            pj := map[string]any{"name": p.Name}
            if len(p.Collect) > 0 {
                cols := make([]any, 0, len(p.Collect))
                for _, c := range p.Collect {
                    cj := map[string]any{"step": c.Step}
                    if c.Merge != nil {
                        mj := map[string]any{}
                        if c.Merge.Stable { mj["stable"] = true }
                        if len(c.Merge.Sort) > 0 {
                            sk := make([]any, 0, len(c.Merge.Sort))
                            for _, s := range c.Merge.Sort { sk = append(sk, map[string]any{"field": s.Field, "order": s.Order}) }
                            mj["sort"] = sk
                        }
                        if c.Merge.Key != "" { mj["key"] = c.Merge.Key }
                        if c.Merge.PartitionBy != "" { mj["partitionBy"] = c.Merge.PartitionBy }
                        if c.Merge.Buffer.Capacity > 0 || c.Merge.Buffer.Policy != "" {
                            mj["buffer"] = map[string]any{"capacity": c.Merge.Buffer.Capacity, "policy": c.Merge.Buffer.Policy}
                        }
                        if c.Merge.Window > 0 { mj["window"] = c.Merge.Window }
                        if c.Merge.TimeoutMs > 0 { mj["timeoutMs"] = c.Merge.TimeoutMs }
                        if c.Merge.DedupField != "" { mj["dedupField"] = c.Merge.DedupField }
                        if c.Merge.Watermark != nil { mj["watermark"] = map[string]any{"field": c.Merge.Watermark.Field, "latenessMs": c.Merge.Watermark.LatenessMs} }
                        if c.Merge.LatePolicy != "" { mj["latePolicy"] = c.Merge.LatePolicy }
                        cj["merge"] = mj
                    }
                    cols = append(cols, cj)
                }
                pj["collect"] = cols
            }
            ps = append(ps, pj)
        }
        jm["pipelines"] = ps
    }
    fns := make([]any, 0, len(m.Functions))
    for _, f := range m.Functions {
        jf := map[string]any{
            "name":    f.Name,
            "params":  valuesToJSON(f.Params),
            "results": valuesToJSON(f.Results),
            "blocks":  []any{},
        }
        if len(f.Decorators) > 0 {
            decs := make([]any, 0, len(f.Decorators))
            for _, d := range f.Decorators {
                obj := map[string]any{"name": d.Name}
                if len(d.Args) > 0 { obj["args"] = d.Args }
                decs = append(decs, obj)
            }
            jf["decorators"] = decs
        }
        bl := make([]any, 0, len(f.Blocks))
        for _, b := range f.Blocks {
            jb := map[string]any{"name": b.Name, "instrs": instrsToJSON(b.Instr)}
            bl = append(bl, jb)
        }
        jf["blocks"] = bl
        if len(f.GPUBlocks) > 0 {
            gbs := make([]any, 0, len(f.GPUBlocks))
            for _, g := range f.GPUBlocks {
                obj := map[string]any{"family": g.Family, "source": g.Source}
                if g.Name != "" { obj["name"] = g.Name }
                if g.N > 0 { obj["n"] = g.N }
                if g.Grid != [3]int{} { obj["grid"] = []int{g.Grid[0], g.Grid[1], g.Grid[2]} }
                if g.TPG != [3]int{} { obj["tpg"] = []int{g.TPG[0], g.TPG[1], g.TPG[2]} }
                if g.Args != "" { obj["args"] = g.Args }
                gbs = append(gbs, obj)
            }
            jf["gpuBlocks"] = gbs
        }
        fns = append(fns, jf)
    }
    jm["functions"] = fns
    return json.MarshalIndent(jm, "", "  ")
}
