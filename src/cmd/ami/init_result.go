package main

type initResult struct {
    Created           bool     `json:"created"`
    Updated           bool     `json:"updated"`
    WorkspacePath     string   `json:"workspacePath"`
    TargetDirCreated  bool     `json:"targetDirCreated"`
    PackageDirCreated bool     `json:"packageDirCreated"`
    GitStatus         string   `json:"gitStatus"` // present|initialized|required
    Messages          []string `json:"messages"`
}

