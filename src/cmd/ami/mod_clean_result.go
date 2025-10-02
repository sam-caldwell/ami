package main

type modCleanResult struct {
    Path    string `json:"path"`
    Removed bool   `json:"removed"`
    Created bool   `json:"created"`
}

