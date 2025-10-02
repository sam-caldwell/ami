package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sort"
)

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

