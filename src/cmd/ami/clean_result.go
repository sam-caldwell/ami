package main

type cleanResult struct {
    Path     string   `json:"path"`
    Removed  bool     `json:"removed"`
    Created  bool     `json:"created"`
    Messages []string `json:"messages"`
}

