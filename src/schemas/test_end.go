package schemas

import "errors"

// TestEnd marks the end of a single test.
type TestEnd struct {
    Schema      string   `json:"schema"`
    Type        string   `json:"type"`
    Timestamp   string   `json:"timestamp"`
    Package     string   `json:"package"`
    Name        string   `json:"name"`
    Status      string   `json:"status"` // pass|fail|skip
    DurationMs  int64    `json:"duration_ms"`
    Diagnostics []DiagV1 `json:"diagnostics,omitempty"`
}

func (e *TestEnd) Validate() error {
    if e.Schema == "" { e.Schema = "test.v1" }
    if e.Schema != "test.v1" { return errors.New("invalid schema") }
    if e.Type == "" { e.Type = "test_end" }
    if e.Type != "test_end" { return errors.New("invalid type") }
    switch e.Status {
    case "pass", "fail", "skip":
    default:
        return errors.New("invalid status")
    }
    return nil
}

