package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "time"
)

// resolvedUnit summarizes a single source unit for the resolved sources listing.
type resolvedUnit struct {
    Package string   `json:"package"`
    File    string   `json:"file"`
    Imports []string `json:"imports,omitempty"`
    Source  string   `json:"source"`
}

// resolvedPayload is the toplevel object for build/debug/source/resolved.json.
type resolvedPayload struct {
    Schema    string          `json:"schema"`
    Timestamp time.Time       `json:"timestamp"`
    Units     []resolvedUnit  `json:"units"`
}

// writeResolvedSourcesDebug writes build/debug/source/resolved.json with a list of units.
func writeResolvedSourcesDebug(units []resolvedUnit) (string, error) {
    dir := filepath.Join("build", "debug", "source")
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    obj := resolvedPayload{Schema: "sources.v1", Timestamp: time.Now().UTC(), Units: units}
    b, err := json.MarshalIndent(obj, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, "resolved.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}

