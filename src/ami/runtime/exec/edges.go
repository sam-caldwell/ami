package exec

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type edgeEntry struct {
    Unit     string `json:"unit"`
    Pipeline string `json:"pipeline"`
    From     string `json:"from"`
    To       string `json:"to"`
}

type edgesIndex struct {
    Schema  string      `json:"schema"`
    Package string      `json:"package"`
    Edges   []edgeEntry `json:"edges"`
}

// BuildLinearPathFromEdges loads build/debug/asm/<pkg>/edges.json and
// constructs a simple ingress→…→egress linear path for the given pipeline.
// If multiple out-edges exist, it picks the first found deterministically.
func BuildLinearPathFromEdges(rootDir, pkg, pipeline string) ([]string, error) {
    path := filepath.Join(rootDir, "build", "debug", "asm", pkg, "edges.json")
    b, err := os.ReadFile(path)
    if err != nil { return nil, err }
    var idx edgesIndex
    if err := json.Unmarshal(b, &idx); err != nil { return nil, err }
    // Build adjacency for the target pipeline
    adj := map[string][]string{}
    nodes := map[string]bool{}
    for _, e := range idx.Edges {
        if e.Pipeline != pipeline { continue }
        nodes[e.From] = true; nodes[e.To] = true
        adj[e.From] = append(adj[e.From], e.To)
    }
    // Construct a simple path from 'ingress' to 'egress'
    pathNodes := []string{"ingress"}
    seen := map[string]bool{"ingress": true}
    cur := "ingress"
    // prevent infinite loops by limiting hops
    for steps := 0; steps < 1_000; steps++ {
        if cur == "egress" { break }
        outs := adj[cur]
        if len(outs) == 0 { break }
        next := outs[0]
        // prefer a direct Collect/Transform/Egress if present in outs deterministically
        // but keep outs[0] to maintain edges.json order
        if !seen[next] { pathNodes = append(pathNodes, next); seen[next] = true; cur = next } else { break }
    }
    return pathNodes, nil
}

