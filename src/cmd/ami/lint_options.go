package main

// LintOptions holds CLI-configured behavior for lint execution.
type LintOptions struct {
    Rules       []string
    MaxWarn     int // -1 means unlimited
    FailFast    bool
    CompatCodes bool
}

var currentLintOptions = LintOptions{MaxWarn: -1}

// setLintOptions applies options. Passing an all-zero LintOptions (i.e.,
// LintOptions{}) resets to defaults (MaxWarn:-1). This matches test usage
// where defer setLintOptions(LintOptions{}) is intended to restore defaults.
// setLintOptions moved to lint_options_set.go
