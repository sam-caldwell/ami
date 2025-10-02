package main

// setLintOptions applies options. Passing an all-zero LintOptions (i.e.,
// LintOptions{}) resets to defaults (MaxWarn:-1). This matches test usage
// where defer setLintOptions(LintOptions{}) is intended to restore defaults.
func setLintOptions(o LintOptions) {
    if len(o.Rules) == 0 && o.MaxWarn == 0 && !o.FailFast && !o.CompatCodes {
        currentLintOptions = LintOptions{MaxWarn: -1}
        return
    }
    currentLintOptions = o
}

