package schemas

// BuildPlanV1 describes the build execution plan.
type BuildPlanV1 struct {
    Schema    string        `json:"schema"`
    Timestamp string        `json:"timestamp"`
    Workspace string        `json:"workspace"`
    Toolchain ToolchainV1   `json:"toolchain"`
    Targets   []BuildTarget `json:"targets"`
    Graph     *BuildGraph   `json:"graph,omitempty"`
}

