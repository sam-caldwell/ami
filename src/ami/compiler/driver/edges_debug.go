package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sort"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

type edgeEntry struct {
    Unit     string `json:"unit"`
    From     string `json:"from"`
    To       string `json:"to"`
    Bounded  bool   `json:"bounded"`
    Delivery string `json:"delivery"`
}

type edgesIndex struct {
    Schema  string      `json:"schema"`
    Package string      `json:"package"`
    Edges   []edgeEntry `json:"edges"`
    Collect []collectEntry `json:"collect,omitempty"`
}

type collectEntry struct {
    Unit      string       `json:"unit"`
    Step      string       `json:"step"`
    MultiPath *edgeMultiPath `json:"multipath,omitempty"`
}

type edgeMultiPath struct {
    Args  []string   `json:"args"`
    Merge []mergeAttr `json:"merge"`
}

type mergeAttr struct {
    Name string   `json:"name"`
    Args []string `json:"args"`
}

// collectEdges returns all edge entries for a parsed file, tagged with unit.
func collectEdges(unit string, f *ast.File) []edgeEntry {
    var out []edgeEntry
    if f == nil { return out }
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        // build a quick index of step attrs by step name
        stepAttrs := map[string][]ast.Attr{}
        for _, s := range pd.Stmts {
            if st, ok := s.(*ast.StepStmt); ok {
                stepAttrs[st.Name] = st.Attrs
            }
        }
        for _, s := range pd.Stmts {
            if e, ok := s.(*ast.EdgeStmt); ok {
                // derive bounded/delivery from target step attributes (scaffold defaults)
                bounded := false
                delivery := "atLeastOnce"
                for _, at := range stepAttrs[e.To] {
                    // simple delivery inference
                    if at.Name == "dropOldest" || at.Name == "dropNewest" { delivery = "bestEffort" }
                }
                out = append(out, edgeEntry{Unit: unit, From: e.From, To: e.To, Bounded: bounded, Delivery: delivery})
            }
        }
    }
    return out
}

// writeEdgesIndex writes the edges index for a package.
func writeEdgesIndex(pkg string, edges []edgeEntry, collects []collectEntry) (string, error) {
    // sort deterministically by unit/from/to
    sort.SliceStable(edges, func(i, j int) bool {
        if edges[i].Unit != edges[j].Unit { return edges[i].Unit < edges[j].Unit }
        if edges[i].From != edges[j].From { return edges[i].From < edges[j].From }
        return edges[i].To < edges[j].To
    })
    // sort collects by unit/step
    sort.SliceStable(collects, func(i, j int) bool {
        if collects[i].Unit != collects[j].Unit { return collects[i].Unit < collects[j].Unit }
        return collects[i].Step < collects[j].Step
    })
    idx := edgesIndex{Schema: "edges.v1", Package: pkg, Edges: edges, Collect: collects}
    dir := filepath.Join("build", "debug", "asm", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(idx, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, "edges.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}
