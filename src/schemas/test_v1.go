package schemas

import "errors"

// Test stream (JSON lines) v1

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

type TestStart struct {
	Schema    string `json:"schema"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Package   string `json:"package"`
	Name      string `json:"name"`
}

type TestOutput struct {
	Schema    string `json:"schema"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Package   string `json:"package"`
	Name      string `json:"name"`
	Stream    string `json:"stream"` // stdout|stderr
	Text      string `json:"text"`
}

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

// Per-package summary reported at run_end
type TestPackageSummary struct {
	Package string `json:"package"`
	Pass    int    `json:"pass"`
	Fail    int    `json:"fail"`
	Skip    int    `json:"skip"`
	Cases   int    `json:"cases"`
}

func (e *TestRunStart) Validate() error {
	if e.Schema == "" {
		e.Schema = "test.v1"
	}
	if e.Schema != "test.v1" {
		return errors.New("invalid schema")
	}
	if e.Type == "" {
		e.Type = "run_start"
	}
	if e.Type != "run_start" {
		return errors.New("invalid type")
	}
	return nil
}
func (e *TestStart) Validate() error {
	if e.Schema == "" {
		e.Schema = "test.v1"
	}
	if e.Schema != "test.v1" {
		return errors.New("invalid schema")
	}
	if e.Type == "" {
		e.Type = "test_start"
	}
	if e.Type != "test_start" {
		return errors.New("invalid type")
	}
	return nil
}
func (e *TestOutput) Validate() error {
	if e.Schema == "" {
		e.Schema = "test.v1"
	}
	if e.Schema != "test.v1" {
		return errors.New("invalid schema")
	}
	if e.Type == "" {
		e.Type = "test_output"
	}
	if e.Type != "test_output" {
		return errors.New("invalid type")
	}
	if e.Stream != "stdout" && e.Stream != "stderr" {
		return errors.New("invalid stream")
	}
	return nil
}
func (e *TestEnd) Validate() error {
	if e.Schema == "" {
		e.Schema = "test.v1"
	}
	if e.Schema != "test.v1" {
		return errors.New("invalid schema")
	}
	if e.Type == "" {
		e.Type = "test_end"
	}
	if e.Type != "test_end" {
		return errors.New("invalid type")
	}
	switch e.Status {
	case "pass", "fail", "skip":
	default:
		return errors.New("invalid status")
	}
	return nil
}
func (e *TestRunEnd) Validate() error {
	if e.Schema == "" {
		e.Schema = "test.v1"
	}
	if e.Schema != "test.v1" {
		return errors.New("invalid schema")
	}
	if e.Type == "" {
		e.Type = "run_end"
	}
	if e.Type != "run_end" {
		return errors.New("invalid type")
	}
	return nil
}
