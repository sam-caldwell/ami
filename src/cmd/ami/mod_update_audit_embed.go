package main

// modAuditEmbed mirrors key fields from AuditReport for JSON embedding in update result.
type modAuditEmbed struct {
    MissingInSum   []string `json:"missingInSum,omitempty"`
    Unsatisfied    []string `json:"unsatisfied,omitempty"`
    MissingInCache []string `json:"missingInCache,omitempty"`
    Mismatched     []string `json:"mismatched,omitempty"`
    ParseErrors    []string `json:"parseErrors,omitempty"`
    SumFound       bool     `json:"sumFound"`
}

