package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sort"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

type contractsDoc struct {
    Schema       string               `json:"schema"`
    Package      string               `json:"package"`
    Unit         string               `json:"unit"`
    Delivery     string               `json:"delivery"`
    Capabilities []string             `json:"capabilities,omitempty"`
    TrustLevel   string               `json:"trustLevel,omitempty"`
    Pipelines    []contractPipeline   `json:"pipelines"`
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
}

// writeContractsDebug writes a minimal contracts.json snapshot with delivery policy,
// capabilities, trust level, and per-step type and edge policy.
func writeContractsDebug(pkg, unit string, f *ast.File) (string, error) {
    defaultDelivery := "atLeastOnce"
    var capabilities []string
    var trustLevel string
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
                        if p != "" { capabilities = append(capabilities, p) }
                    }
                }
                if len(pr.Args) > 0 { capabilities = append(capabilities, pr.Args...) }
            case "trust":
                if lv, ok := pr.Params["level"]; ok { trustLevel = lv }
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
                    if strings.HasPrefix(st.Name, "io.") {
                        // broad mapping
                        head := strings.TrimPrefix(st.Name, "io.")
                        head = strings.ToLower(head)
                        if strings.HasPrefix(head, "read") || strings.HasPrefix(head, "recv") { addCap("io.read") }
                        if strings.HasPrefix(head, "write") || strings.HasPrefix(head, "send") { addCap("io.write") }
                        if strings.HasPrefix(head, "connect") || strings.HasPrefix(head, "listen") || strings.HasPrefix(head, "dial") { addCap("network") }
                    }
                    steps = append(steps, contractStep{Name: st.Name, Type: typ, Bounded: bounded, Delivery: del})
                }
            }
            pentries = append(pentries, contractPipeline{Name: pd.Name, Steps: steps})
        }
    }
    sort.Strings(capabilities)
    sort.SliceStable(pentries, func(i, j int) bool { return pentries[i].Name < pentries[j].Name })
    obj := contractsDoc{Schema: "contracts.v1", Package: pkg, Unit: unit, Delivery: defaultDelivery, Capabilities: capabilities, TrustLevel: trustLevel, Pipelines: pentries}
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(obj, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, unit+".contracts.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}

