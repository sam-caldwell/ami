package driver

type edgeAttrs struct {
    Bounded  bool   `json:"bounded"`
    Delivery string `json:"delivery"`
    Type     string `json:"type,omitempty"`
}

