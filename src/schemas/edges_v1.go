package schemas

import "errors"

// EdgesV1 summarizes compiler-discovered input edges per package
// for debug/inspection. Emitted under build/debug/asm/<package>/edges.json
// when `ami build --verbose` is used.
type EdgesV1 struct {
    Schema    string       `json:"schema"`
    Timestamp string       `json:"timestamp"`
    Package   string       `json:"package"`
    Items     []EdgeInitV1 `json:"items"`
}

type EdgeInitV1 struct {
    Unit         string `json:"unit"`
    Pipeline     string `json:"pipeline"`
    Segment      string `json:"segment"` // normal|error
    Step         int    `json:"step"`
    Node         string `json:"node"`
    Label        string `json:"label"`
    Kind         string `json:"kind"`
    MinCapacity  int    `json:"minCapacity,omitempty"`
    MaxCapacity  int    `json:"maxCapacity,omitempty"`
    Backpressure string `json:"backpressure,omitempty"`
    Type         string `json:"type,omitempty"`
    UpstreamName string `json:"upstreamName,omitempty"`
    // Derived semantics
    Bounded      bool   `json:"bounded,omitempty"`   // true when maxCapacity > 0; false indicates unbounded
    Delivery     string `json:"delivery,omitempty"`  // atLeastOnce (block) | bestEffort (drop)
}

func (e *EdgesV1) Validate() error {
    if e == nil { return errors.New("nil edges") }
    if e.Schema == "" { e.Schema = "edges.v1" }
    if e.Schema != "edges.v1" { return errors.New("invalid schema") }
    return nil
}
