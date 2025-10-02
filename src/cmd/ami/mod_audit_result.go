package main

import "github.com/sam-caldwell/ami/src/ami/workspace"

type modAuditResult struct {
    Requirements   []workspace.Requirement `json:"requirements"`
    MissingInSum   []string                `json:"missingInSum"`
    Unsatisfied    []string                `json:"unsatisfied"`
    MissingInCache []string                `json:"missingInCache"`
    Mismatched     []string                `json:"mismatched"`
    ParseErrors    []string                `json:"parseErrors"`
    SumFound       bool                    `json:"sumFound"`
    Timestamp      string                  `json:"timestamp"`
}

