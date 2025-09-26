package logger

// record is the JSON schema envelope written in JSON mode.
type record struct {
    Schema    string                 `json:"schema"`
    Timestamp string                 `json:"timestamp"`
    Level     string                 `json:"level"`
    Message   string                 `json:"message"`
    Data      map[string]interface{} `json:"data,omitempty"`
}

