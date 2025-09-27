package main

// LintOptions holds CLI-configured behavior for lint execution.
type LintOptions struct {
    Rules       []string
    MaxWarn     int // -1 means unlimited
    FailFast    bool
    CompatCodes bool
}

var currentLintOptions = LintOptions{MaxWarn: -1}

func setLintOptions(o LintOptions) { currentLintOptions = o }

