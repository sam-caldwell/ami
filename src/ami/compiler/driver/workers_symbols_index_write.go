package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "sort"
)

func writeWorkersSymbolsIndex(pkg string, workers []string) (string, error) {
    if len(workers) == 0 { return "", nil }
    uniq := map[string]struct{}{}
    var list []string
    for _, w := range workers { if w != "" { if _, ok := uniq[w]; !ok { uniq[w] = struct{}{}; list = append(list, w) } } }
    sort.Strings(list)
    var idx workersSymbolsIndex
    idx.Schema = "workers.symbols.v1"
    idx.Package = pkg
    for _, w := range list {
        s := sanitizeForCSymbol("ami_worker_", w)
        impl := sanitizeForCSymbol("ami_worker_impl_", w)
        idx.Symbols = append(idx.Symbols, struct{
            Name      string `json:"name"`
            Sanitized string `json:"sanitized"`
            Impl      string `json:"impl"`
        }{Name: w, Sanitized: s, Impl: impl})
    }
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(idx, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, "workers.symbols.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}
