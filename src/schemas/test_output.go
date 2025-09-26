package schemas

import "errors"

// TestOutput captures a line of test output.
type TestOutput struct {
    Schema    string `json:"schema"`
    Type      string `json:"type"`
    Timestamp string `json:"timestamp"`
    Package   string `json:"package"`
    Name      string `json:"name"`
    Stream    string `json:"stream"` // stdout|stderr
    Text      string `json:"text"`
}

func (e *TestOutput) Validate() error {
    if e.Schema == "" { e.Schema = "test.v1" }
    if e.Schema != "test.v1" { return errors.New("invalid schema") }
    if e.Type == "" { e.Type = "test_output" }
    if e.Type != "test_output" { return errors.New("invalid type") }
    if e.Stream != "stdout" && e.Stream != "stderr" { return errors.New("invalid stream") }
    return nil
}

