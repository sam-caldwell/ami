package main

type modGetResult struct {
    Source  string `json:"source"`
    Name    string `json:"name"`
    Version string `json:"version"`
    Path    string `json:"path"`
    Message string `json:"message,omitempty"`
}

