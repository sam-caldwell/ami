package schemas

import "errors"

type SourcesV1 struct {
	Schema    string       `json:"schema"`
	Timestamp string       `json:"timestamp"`
	Units     []SourceUnit `json:"units"`
}

type SourceUnit struct {
	Package         string         `json:"package"`
	File            string         `json:"file"`
	Imports         []string       `json:"imports"`
	ImportsDetailed []SourceImport `json:"importsDetailed,omitempty"`
	Source          string         `json:"source"`
}

// SourceImport describes a single import with optional alias and constraint.
type SourceImport struct {
	Path       string `json:"path"`
	Alias      string `json:"alias,omitempty"`
	Constraint string `json:"constraint,omitempty"`
}

func (s *SourcesV1) Validate() error {
	if s == nil {
		return errors.New("nil sources")
	}
	if s.Schema == "" {
		s.Schema = "sources.v1"
	}
	if s.Schema != "sources.v1" {
		return errors.New("invalid schema")
	}
	return nil
}
