package driver

type contractStep struct {
    Name      string `json:"name"`
    Type      string `json:"type,omitempty"`
    Bounded   bool   `json:"bounded"`
    Delivery  string `json:"delivery"`
    MinCapacity int   `json:"minCapacity,omitempty"`
    MaxCapacity int   `json:"maxCapacity,omitempty"`
    Backpressure string `json:"backpressure,omitempty"`
    ExecModel string `json:"execModel,omitempty"`
}
