package schemas

import "errors"

// TestRunEnd summarizes a test run.
type TestRunEnd struct {
    Schema     string `json:"schema"`
    Type       string `json:"type"`
    Timestamp  string `json:"timestamp"`
    DurationMs int64  `json:"duration_ms"`
    Totals     struct {
        Pass  int `json:"pass"`
        Fail  int `json:"fail"`
        Skip  int `json:"skip"`
        Cases int `json:"cases"`
    } `json:"totals"`
    Packages []TestPackageSummary `json:"packages,omitempty"`
}

func (e *TestRunEnd) Validate() error {
    if e.Schema == "" { e.Schema = "test.v1" }
    if e.Schema != "test.v1" { return errors.New("invalid schema") }
    if e.Type == "" { e.Type = "run_end" }
    if e.Type != "run_end" { return errors.New("invalid type") }
    return nil
}

