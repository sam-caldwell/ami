package main

type modListEntry struct {
    Name     string `json:"name"`
    Version  string `json:"version,omitempty"`
    Type     string `json:"type"`   // file|dir
    Size     int64  `json:"size"`
    Modified string `json:"modified"` // ISO-8601 UTC
}

