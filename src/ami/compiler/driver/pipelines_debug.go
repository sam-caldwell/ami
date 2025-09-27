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
    Merge []mergeAttr `json:"merge,omitempty"`
    MultiPath *multiPath `json:"multipath,omitempty"`
}

type edgeAttrs struct {
    Bounded  bool   `json:"bounded"`
    Delivery string `json:"delivery"`
}

type mergeAttr struct {
    Name string   `json:"name"`
    Args []string `json:"args"`
}

type multiPath struct {
    Args []string `json:"args"`
}

// writePipelinesDebug writes pipelines debug JSON for a parsed file.
func writePipelinesDebug(pkg, unit string, f *ast.File) (string, error) {
    var entries []pipelineEntry
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
                op.Edge = &edgeAttrs{Bounded: false, Delivery: "atLeastOnce"}
                // merge.* attributes captured verbatim (scaffold)
                for _, at := range st.Attrs {
                    if len(at.Name) >= 6 && at.Name[:6] == "merge." {
                        var margs []string
                        for _, aa := range at.Args { margs = append(margs, aa.Text) }
                        op.Merge = append(op.Merge, mergeAttr{Name: at.Name, Args: margs})
                    }
                    if (at.Name == "edge.MultiPath" || at.Name == "MultiPath") && st.Name == "Collect" {
                        var margs []string
                        for _, aa := range at.Args { margs = append(margs, aa.Text) }
                        op.MultiPath = &multiPath{Args: margs}
                    }
                }
                steps = append(steps, op)
            }
        }
        entries = append(entries, pipelineEntry{Name: pd.Name, Steps: steps})
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
