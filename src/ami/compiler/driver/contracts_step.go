package driver

type contractStep struct {
    Name      string `json:"name"`
    Type      string `json:"type,omitempty"`
    Bounded   bool   `json:"bounded"`
    Delivery  string `json:"delivery"`
    ExecModel string `json:"execModel,omitempty"`
}

