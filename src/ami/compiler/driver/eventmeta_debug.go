package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type eventMeta struct {
    Schema           string         `json:"schema"`
    Package          string         `json:"package"`
    Unit             string         `json:"unit"`
    ID               string         `json:"id"`
    Timestamp        string         `json:"timestamp"`
    Attempt          int            `json:"attempt"`
    Trace            traceMeta      `json:"trace"`
    ImmutablePayload bool           `json:"immutablePayload"`
}
type traceMeta struct {
    Traceparent string `json:"traceparent"`
    Tracestate  string `json:"tracestate"`
}

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

