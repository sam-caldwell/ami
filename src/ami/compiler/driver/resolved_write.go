package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "time"
)

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

