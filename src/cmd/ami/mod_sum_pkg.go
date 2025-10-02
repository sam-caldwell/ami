package main

type modSumPkg struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    Sha256  string `json:"sha256"`
    Source  string `json:"source,omitempty"`
}

