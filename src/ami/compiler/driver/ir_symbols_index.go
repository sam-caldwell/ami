package driver

import (
    "encoding/json"
    "os"
    "path/filepath"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

type irSymbolsIndexUnit struct {
    Unit    string   `json:"unit"`
    Exports []string `json:"exports,omitempty"`
    Externs []string `json:"externs,omitempty"`
}

type irSymbolsIndex struct {
    Schema  string               `json:"schema"`
    Package string               `json:"package"`
    Units   []irSymbolsIndexUnit `json:"units"`
}

func collectExports(m ir.Module) []string {
    var names []string
    for _, f := range m.Functions { names = append(names, f.Name) }
    sortStrings(names)
    return names
}

// writeIRSymbolsIndex writes ir.symbols.index.json for a package.
func writeIRSymbolsIndex(pkg string, units []irSymbolsIndexUnit) (string, error) {
    idx := irSymbolsIndex{Schema: "ir.symbols.index.v1", Package: pkg, Units: units}
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(idx, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, "ir.symbols.index.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}

