package schemas

import (
	"errors"
)

// DiagV1 implements the JSON diagnostic schema diag.v1
type DiagV1 struct {
	Schema    string                 `json:"schema"`
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"` // info|warn|error
	Code      string                 `json:"code,omitempty"`
	Message   string                 `json:"message"`
	Package   string                 `json:"package,omitempty"`
	File      string                 `json:"file,omitempty"`
	Pos       *Position              `json:"pos,omitempty"`
	EndPos    *Position              `json:"endPos,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Offset int `json:"offset"`
}

func (d *DiagV1) Validate() error {
	if d == nil {
		return errors.New("nil diag")
	}
	if d.Schema == "" {
		d.Schema = "diag.v1"
	}
	if d.Schema != "diag.v1" {
		return errors.New("invalid schema")
	}
	switch d.Level {
	case "info", "warn", "error":
	default:
		return errors.New("invalid level")
	}
	if d.Message == "" {
		return errors.New("message required")
	}
	return nil
}
