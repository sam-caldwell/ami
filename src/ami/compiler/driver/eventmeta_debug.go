package driver

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
// trace type and writer moved to separate files to satisfy single-declaration rule
