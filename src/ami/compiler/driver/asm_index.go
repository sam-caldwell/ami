package driver

import ()

type asmIndex struct {
    Schema   string        `json:"schema"`
    Package  string        `json:"package"`
    Tiny     []asmEdge     `json:"tiny,omitempty"`
    Types    []string      `json:"types,omitempty"`
    Policy   map[string]int `json:"policyCount,omitempty"`
    Bounded  int           `json:"boundedCount,omitempty"`
    Total    int           `json:"totalEdges"`
}

// asmEdge and writeAsmIndex moved to separate files to satisfy single-declaration rule
