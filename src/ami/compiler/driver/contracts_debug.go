package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "strconv"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

type contractsDoc struct {
    Schema       string               `json:"schema"`
    Package      string               `json:"package"`
    Unit         string               `json:"unit"`
    Delivery     string               `json:"delivery"`
    Capabilities []string             `json:"capabilities,omitempty"`
    TrustLevel   string               `json:"trustLevel,omitempty"`
    Concurrency  *contractConcurrency  `json:"concurrency,omitempty"`
    Pipelines    []contractPipeline   `json:"pipelines"`
    CapabilityNotes []string          `json:"capabilityNotes,omitempty"`
    SchemaStability string            `json:"schemaStability,omitempty"`
    SchemaNotes   []string            `json:"schemaNotes,omitempty"`
}

type contractPipeline struct {
    Name  string            `json:"name"`
    Steps []contractStep    `json:"steps"`
}

type contractStep struct {
    Name     string         `json:"name"`
    Type     string         `json:"type,omitempty"`
    Bounded  bool           `json:"bounded"`
    Delivery string         `json:"delivery"`
    ExecModel string        `json:"execModel,omitempty"`
}

type contractConcurrency struct {
    Workers  int    `json:"workers,omitempty"`
    Schedule string `json:"schedule,omitempty"`
    Limits   map[string]int `json:"limits,omitempty"`
}

// writeContractsDebug writes a minimal contracts.json snapshot with delivery policy,
// capabilities, trust level, and per-step type and edge policy.
func writeContractsDebug(pkg, unit string, f *ast.File) (string, error) {
    defaultDelivery := "atLeastOnce"
    var capabilities []string
    declaredCaps := map[string]bool{}
    var trustLevel string
    var conc *contractConcurrency
    if f != nil {
        for _, pr := range f.Pragmas {
            switch pr.Domain {
            case "backpressure":
                if pol, ok := pr.Params["policy"]; ok {
                    switch pol {
                    case "dropOldest", "dropNewest":
                        defaultDelivery = "bestEffort"
                    case "block":
                        defaultDelivery = "atLeastOnce"
                    }
                }
            case "capabilities":
                if lst, ok := pr.Params["list"]; ok && lst != "" {
                    for _, p := range strings.Split(lst, ",") {
                        p = strings.TrimSpace(p)
                        if p != "" {
                            capabilities = append(capabilities, p)
                            declaredCaps[strings.ToLower(p)] = true
                        }
                    }
                }
                if len(pr.Args) > 0 {
                    capabilities = append(capabilities, pr.Args...)
                    for _, a := range pr.Args { if a != "" { declaredCaps[strings.ToLower(a)] = true } }
                }
            case "trust":
                if lv, ok := pr.Params["level"]; ok { trustLevel = lv }
            case "concurrency":
                if pr.Key == "workers" {
                    if n, ok := pr.Params["count"]; ok && n != "" {
                        if conc == nil { conc = &contractConcurrency{} }
                        // best-effort parse; semantics pass validates
                        if w, err := strconv.Atoi(n); err == nil { conc.Workers = w }
                    } else if pr.Value != "" {
                        if conc == nil { conc = &contractConcurrency{} }
                        if w, err := strconv.Atoi(pr.Value); err == nil { conc.Workers = w }
                    }
                }
                if pr.Key == "schedule" {
                    if conc == nil { conc = &contractConcurrency{} }
                    conc.Schedule = pr.Value
                }
                if pr.Key == "limits" {
                    if conc == nil { conc = &contractConcurrency{} }
                    if conc.Limits == nil { conc.Limits = map[string]int{} }
                    // collect params and args k=v
                    for k, v := range pr.Params {
                        if w, err := strconv.Atoi(v); err == nil { conc.Limits[k] = w }
                    }
                    for _, a := range pr.Args {
                        if eq := strings.IndexByte(a, '='); eq > 0 {
                            k := a[:eq]
                            v := a[eq+1:]
                            if w, err := strconv.Atoi(v); err == nil { conc.Limits[k] = w }
                        }
                    }
                }
            }
        }
    }
    // Derive additional capabilities from io.* steps
    addCap := func(cap string) {
        if cap == "" { return }
        for _, c := range capabilities { if c == cap { return } }
        capabilities = append(capabilities, cap)
    }
    var pentries []contractPipeline
    var sawIOReadWrite bool
    if f != nil {
        for _, d := range f.Decls {
            pd, ok := d.(*ast.PipelineDecl)
            if !ok { continue }
            var steps []contractStep
            for _, s := range pd.Stmts {
                if st, ok := s.(*ast.StepStmt); ok {
                    // type from attrs if present
                    typ := ""
                    bounded := false
                    del := defaultDelivery
                    exec := "thread"
                    for _, at := range st.Attrs {
                        if (at.Name == "type" || at.Name == "Type") && len(at.Args) > 0 && at.Args[0].Text != "" {
                            typ = at.Args[0].Text
                        }
                        if len(at.Name) >= 6 && at.Name[:6] == "merge." {
                            if at.Name == "merge.Buffer" {
                                if len(at.Args) > 0 && at.Args[0].Text != "0" && at.Args[0].Text != "" { bounded = true }
                                if len(at.Args) > 1 {
                                    pol := at.Args[1].Text
                                    if pol == "dropOldest" || pol == "dropNewest" { del = "bestEffort" }
                                    if pol == "block" { del = "atLeastOnce" }
                                }
                            }
                        }
                    }
                    lname := strings.ToLower(st.Name)
                    if strings.HasPrefix(lname, "io.") {
                        // broad mapping
                        head := strings.TrimPrefix(lname, "io.")
                        head = strings.ToLower(head)
                        if strings.HasPrefix(head, "read") || strings.HasPrefix(head, "recv") { addCap("io.read"); sawIOReadWrite = true }
                        if strings.HasPrefix(head, "write") || strings.HasPrefix(head, "send") { addCap("io.write"); sawIOReadWrite = true }
                        if strings.HasPrefix(head, "connect") || strings.HasPrefix(head, "listen") || strings.HasPrefix(head, "dial") { addCap("network") }
                        exec = "process"
                    }
                    if strings.HasPrefix(lname, "net.") {
                        addCap("net")
                        exec = "process"
                    }
                    // Trust-level influence: under untrusted, prefer process isolation
                    if strings.EqualFold(trustLevel, "untrusted") {
                        exec = "process"
                    }
                    steps = append(steps, contractStep{Name: st.Name, Type: typ, Bounded: bounded, Delivery: del, ExecModel: exec})
                }
            }
            pentries = append(pentries, contractPipeline{Name: pd.Name, Steps: steps})
        }
    }
    // Normalize caps: ensure generic 'io' is present if specifics exist.
    hasIO := false
    hasIOSpecific := false
    for _, c := range capabilities { if c == "io" { hasIO = true } }
    for _, c := range capabilities { if strings.HasPrefix(c, "io.") { hasIOSpecific = true; break } }
    if hasIOSpecific && !hasIO { capabilities = append(capabilities, "io") }
    sort.Strings(capabilities)
    sort.SliceStable(pentries, func(i, j int) bool { return pentries[i].Name < pentries[j].Name })
    var notes []string
    if hasIOSpecific { notes = append(notes, "specific io.* capabilities present; generic 'io' implies all io.*") }
    // Suggest specific capabilities when only generic 'io' was declared but read/write used
    if sawIOReadWrite && declaredCaps["io"] && !declaredCaps["io.read"] && !declaredCaps["io.write"] {
        notes = append(notes, "io.read/write operations detected; consider declaring specific capabilities 'io.read'/'io.write' instead of generic 'io'")
    }
    obj := contractsDoc{Schema: "contracts.v1", Package: pkg, Unit: unit, Delivery: defaultDelivery, Capabilities: capabilities, TrustLevel: trustLevel, Concurrency: conc, Pipelines: pentries, CapabilityNotes: notes}
    // Document stability window for the schema to aid consumers; conservative default.
    obj.SchemaStability = "beta"
    obj.SchemaNotes = append(obj.SchemaNotes, "contracts.v1 is stable for debug/analysis; fields may expand in minor milestones without breaking existing keys")
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(obj, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, unit+".contracts.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}
