package exec

type edgeEntry struct {
    Unit     string `json:"unit"`
    Pipeline string `json:"pipeline"`
    From     string `json:"from"`
    To       string `json:"to"`
    // Optional fields for enriched edges
    Bounded  bool   `json:"bounded,omitempty"`
    Delivery string `json:"delivery,omitempty"`
    Type     string `json:"type,omitempty"`
    Tiny     bool   `json:"tinyBuffer,omitempty"`
    MinCapacity int `json:"minCapacity,omitempty"`
    MaxCapacity int `json:"maxCapacity,omitempty"`
    Backpressure string `json:"backpressure,omitempty"`
}
