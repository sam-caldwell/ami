package main

// goTestEvent represents a single event from `go test -json` output.
// Fields match the upstream JSON schema used by the Go toolchain.
type goTestEvent struct {
    Time    string `json:"Time"`
    Action  string `json:"Action"`
    Package string `json:"Package"`
    Test    string `json:"Test,omitempty"`
    Output  string `json:"Output,omitempty"`
}

