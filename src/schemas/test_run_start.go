package schemas

import "errors"

// TestRunStart begins a test run stream.
type TestRunStart struct {
    Schema      string   `json:"schema"`
    Type        string   `json:"type"`
    Timestamp   string   `json:"timestamp"`
    Workspace   string   `json:"workspace"`
    Packages    []string `json:"packages"`
    Timeout     string   `json:"timeout,omitempty"`
    Parallel    int      `json:"parallel,omitempty"`
    PkgParallel int      `json:"pkg_parallel,omitempty"`
    FailFast    bool     `json:"failfast,omitempty"`
    Run         string   `json:"run,omitempty"`
}

func (e *TestRunStart) Validate() error {
    if e.Schema == "" { e.Schema = "test.v1" }
    if e.Schema != "test.v1" { return errors.New("invalid schema") }
    if e.Type == "" { e.Type = "run_start" }
    if e.Type != "run_start" { return errors.New("invalid type") }
    return nil
}

