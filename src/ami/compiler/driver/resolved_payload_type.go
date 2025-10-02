package driver

import "time"

type resolvedPayload struct {
    Schema    string          `json:"schema"`
    Timestamp time.Time       `json:"timestamp"`
    Units     []resolvedUnit  `json:"units"`
}

