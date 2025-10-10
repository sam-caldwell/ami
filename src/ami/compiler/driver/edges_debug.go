package driver

type edgeEntry struct {
    Unit     string `json:"unit"`
    Pipeline string `json:"pipeline"`
    From     string `json:"from"`
    To       string `json:"to"`
    FromID   int    `json:"fromId"`
    ToID     int    `json:"toId"`
    Bounded  bool   `json:"bounded"`
    Delivery string `json:"delivery"`
    Type     string `json:"type,omitempty"`
    Tiny     bool   `json:"tinyBuffer,omitempty"`
    // Explicit capacity/backpressure for runtime consumption
    MinCapacity int    `json:"minCapacity,omitempty"`
    MaxCapacity int    `json:"maxCapacity,omitempty"`
    Backpressure string `json:"backpressure,omitempty"`
    // Derived connectivity hints
    FromReachable bool `json:"fromReachableFromIngress,omitempty"`
    ToReachable   bool `json:"toCanReachEgress,omitempty"`
    OnPath        bool `json:"onIngressToEgressPath,omitempty"`
}
// others moved to separate files to satisfy single-declaration rule
