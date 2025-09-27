package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sort"
)

type asmIndex struct {
    Schema   string        `json:"schema"`
    Package  string        `json:"package"`
    Tiny     []asmEdge     `json:"tiny,omitempty"`
    Types    []string      `json:"types,omitempty"`
    Policy   map[string]int `json:"policyCount,omitempty"`
    Bounded  int           `json:"boundedCount,omitempty"`
    Total    int           `json:"totalEdges"`
}

type asmEdge struct {
    Unit string `json:"unit"`
    From string `json:"from"`
    To   string `json:"to"`
}

// writeAsmIndex writes a summary asm.v1 index for the package.
func writeAsmIndex(pkg string, edges []edgeEntry) (string, error) {
    // collect tiny edges and unique types
    var tiny []asmEdge
    types := map[string]struct{}{}
    for _, e := range edges {
        if e.Tiny { tiny = append(tiny, asmEdge{Unit: e.Unit, From: e.From, To: e.To}) }
        if e.Type != "" { types[e.Type] = struct{}{} }
    }
    // sort tiny edges deterministically
    sort.SliceStable(tiny, func(i, j int) bool {
        if tiny[i].Unit != tiny[j].Unit { return tiny[i].Unit < tiny[j].Unit }
        if tiny[i].From != tiny[j].From { return tiny[i].From < tiny[j].From }
        return tiny[i].To < tiny[j].To
    })
    // sort types
    typeList := make([]string, 0, len(types))
    for k := range types { typeList = append(typeList, k) }
    sort.Strings(typeList)
    // policy counts and bounded/total
    policies := map[string]int{}
    bounded := 0
    for _, e := range edges {
        if e.Delivery != "" { policies[e.Delivery]++ }
        if e.Bounded { bounded++ }
    }
    idx := asmIndex{Schema: "asm.v1", Package: pkg, Tiny: tiny, Types: typeList, Policy: policies, Bounded: bounded, Total: len(edges)}
    dir := filepath.Join("build", "debug", "asm", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(idx, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, "asm.index.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}
