package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
)

// writeEventMetaDebug writes a scaffold event metadata JSON for the unit.
func writeEventMetaDebug(pkg, unit string) (string, error) {
    obj := eventMeta{
        Schema:           "eventmeta.v1",
        Package:          pkg,
        Unit:             unit,
        ID:               "",
        Timestamp:        "1970-01-01T00:00:00.000Z",
        Attempt:          0,
        Trace:            traceMeta{Traceparent: "", Tracestate: ""},
        ImmutablePayload: true,
    }
    dir := filepath.Join("build", "debug", "ir", pkg)
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    b, err := json.MarshalIndent(obj, "", "  ")
    if err != nil { return "", err }
    out := filepath.Join(dir, unit+".eventmeta.json")
    if err := os.WriteFile(out, b, 0o644); err != nil { return "", err }
    return out, nil
}

