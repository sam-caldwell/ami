package schemas

import "errors"

type BuildPlanV1 struct {
	Schema    string        `json:"schema"`
	Timestamp string        `json:"timestamp"`
	Workspace string        `json:"workspace"`
	Toolchain ToolchainV1   `json:"toolchain"`
	Targets   []BuildTarget `json:"targets"`
	Graph     *BuildGraph   `json:"graph,omitempty"`
}

type ToolchainV1 struct {
	AmiVersion string `json:"amiVersion"`
	GoVersion  string `json:"goVersion"`
}

type BuildTarget struct {
	Package string   `json:"package"`
	Unit    string   `json:"unit"`
	Inputs  []string `json:"inputs"`
	Outputs []string `json:"outputs"`
	Steps   []string `json:"steps"`
}

type BuildGraph struct {
	Nodes []string    `json:"nodes"`
	Edges [][2]string `json:"edges"`
}

func (b *BuildPlanV1) Validate() error {
	if b == nil {
		return errors.New("nil build plan")
	}
	if b.Schema == "" {
		b.Schema = "buildplan.v1"
	}
	if b.Schema != "buildplan.v1" {
		return errors.New("invalid schema")
	}
	if b.Workspace == "" {
		return errors.New("workspace required")
	}
	return nil
}
