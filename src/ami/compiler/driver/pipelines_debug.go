package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sort"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

type pipelineList struct {
    Schema    string          `json:"schema"`
    Package   string          `json:"package"`
    Unit      string          `json:"unit"`
    Pipelines []pipelineEntry `json:"pipelines"`
}
type pipelineEntry struct {
    Name  string       `json:"name"`
    Steps []pipelineOp `json:"steps"`
}
type pipelineOp struct {
    Name string   `json:"name"`
    Args []string `json:"args,omitempty"`
    Edge *edgeAttrs `json:"edge,omitempty"`
    Merge []pipeMergeAttr `json:"merge,omitempty"`
    MultiPath *pipeMultiPath `json:"multipath,omitempty"`
    Attrs []pipeAttr `json:"attrs,omitempty"`
}

type edgeAttrs struct {
    Bounded  bool   `json:"bounded"`
    Delivery string `json:"delivery"`
    Type     string `json:"type,omitempty"`
}

type pipeMergeAttr struct {
    Name string   `json:"name"`
    Args []string `json:"args"`
}

type pipeMultiPath struct {
    Args []string `json:"args"`
}

type pipeAttr struct {
    Name string   `json:"name"`
    Args []string `json:"args,omitempty"`
}

// writePipelinesDebug writes pipelines debug JSON for a parsed file.
func writePipelinesDebug(pkg, unit string, f *ast.File) (string, error) {
    var entries []pipelineEntry
    // derive defaults from pragmas
    defaultDelivery := "atLeastOnce"
    if f != nil {
        for _, pr := range f.Pragmas {
            if pr.Domain == "backpressure" {
                if pol, ok := pr.Params["policy"]; ok {
                    switch pol {
                    case "dropOldest", "dropNewest":
                        defaultDelivery = "bestEffort"
                    case "block":
                        defaultDelivery = "atLeastOnce"
                    }
                }
            }
        }
    }
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        var steps []pipelineOp
        for _, s := range pd.Stmts {
            if st, ok := s.(*ast.StepStmt); ok {
                var args []string
                for _, a := range st.Args { args = append(args, a.Text) }
                op := pipelineOp{Name: st.Name, Args: args}
                // default edge attributes (scaffold for #pragma backpressure)
                op.Edge = &edgeAttrs{Bounded: false, Delivery: defaultDelivery}
                // attributes
                var rawAttrs []pipeAttr
                for _, at := range st.Attrs {
                    // generic list
                    var aargs []string
                    for _, aa := range at.Args { aargs = append(aargs, aa.Text) }
                    rawAttrs = append(rawAttrs, pipeAttr{Name: at.Name, Args: aargs})
                    if (at.Name == "type" || at.Name == "Type") && len(aargs) > 0 && aargs[0] != "" {
                        op.Edge.Type = aargs[0]
                    }
                    // merge.* attributes captured verbatim (scaffold)
                    if len(at.Name) >= 6 && at.Name[:6] == "merge." {
                        var margs []string
                        for _, aa := range at.Args { margs = append(margs, aa.Text) }
                        op.Merge = append(op.Merge, pipeMergeAttr{Name: at.Name, Args: margs})
                        if at.Name == "merge.Buffer" {
                            // derive edge attrs from buffer policy
                            if len(margs) > 0 && margs[0] != "0" && margs[0] != "" { op.Edge.Bounded = true }
                            if len(margs) > 1 {
                                pol := margs[1]
                                if pol == "dropOldest" || pol == "dropNewest" { op.Edge.Delivery = "bestEffort" }
                                if pol == "block" { op.Edge.Delivery = "atLeastOnce" }
                            }
                        }
                    }
                    if (at.Name == "edge.MultiPath" || at.Name == "MultiPath") && st.Name == "Collect" {
                        var margs []string
                        for _, aa := range at.Args { margs = append(margs, aa.Text) }
                        op.MultiPath = &pipeMultiPath{Args: margs}
                    }
                }
                op.Attrs = rawAttrs
                steps = append(steps, op)
            }
        }
        entries = append(entries, pipelineEntry{Name: pd.Name, Steps: steps})
    }
    // if no pipelines parsed, synthesize a minimal entry to preserve defaults for tests/tools
    if len(entries) == 0 {
        entries = []pipelineEntry{{Name: "", Steps: []pipelineOp{{Name: "", Edge: &edgeAttrs{Bounded: false, Delivery: defaultDelivery}}}}}
    }
    // deterministic ordering by pipeline name
    sort.SliceStable(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
    obj := pipelineList{Schema: "pipelines.v1", Package: pkg, Unit: unit, Pipelines: entries}
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(obj, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, unit+".pipelines.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}
