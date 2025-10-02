package main

type modListResult struct {
    Path    string         `json:"path"`
    Entries []modListEntry `json:"entries"`
}

