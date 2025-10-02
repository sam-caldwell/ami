package main

type modSumResult struct {
    Path         string   `json:"path,omitempty"`
    Ok           bool     `json:"ok"`
    PackagesSeen int      `json:"packages"`
    Schema       string   `json:"schema"`
    Verified     []string `json:"verified,omitempty"`
    Missing      []string `json:"missing,omitempty"`
    Mismatched   []string `json:"mismatched,omitempty"`
    Message      string   `json:"message,omitempty"`
}

