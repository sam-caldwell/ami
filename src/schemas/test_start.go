package schemas

import "errors"

// TestStart denotes the start of a single test.
type TestStart struct {
    Schema    string `json:"schema"`
    Type      string `json:"type"`
    Timestamp string `json:"timestamp"`
    Package   string `json:"package"`
    Name      string `json:"name"`
}

func (e *TestStart) Validate() error {
    if e.Schema == "" { e.Schema = "test.v1" }
    if e.Schema != "test.v1" { return errors.New("invalid schema") }
    if e.Type == "" { e.Type = "test_start" }
    if e.Type != "test_start" { return errors.New("invalid type") }
    return nil
}

