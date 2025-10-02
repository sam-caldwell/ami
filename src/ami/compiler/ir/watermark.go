package ir

type Watermark struct {
    Field      string `json:"field"`
    LatenessMs int    `json:"latenessMs,omitempty"`
}

