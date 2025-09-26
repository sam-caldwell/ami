package root

import (
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
)

func evalAmiExpectations(c amiCase, diags []diag.Diagnostic) (bool, []diag.Diagnostic) {
    // Build helper maps
    hasErr := false
    errCount := 0
    warnCount := 0
    for _, d := range diags {
        switch d.Level {
        case diag.Error:
            hasErr = true
            errCount++
        case diag.Warn:
            warnCount++
        }
    }
    // All expectations must pass
    for _, e := range c.expects {
        switch e.kind {
        case "no_errors":
            if hasErr {
                // collect errors as failure diagnostics
                var out []diag.Diagnostic
                for _, d := range diags {
                    if d.Level == diag.Error {
                        out = append(out, d)
                    }
                }
                return false, out
            }
        case "no_warnings":
            if warnCount > 0 {
                var out []diag.Diagnostic
                for _, d := range diags {
                    if d.Level == diag.Warn {
                        out = append(out, d)
                    }
                }
                return false, out
            }
        case "error":
            matches := 0
            for _, d := range diags {
                if d.Level == diag.Error && d.Code == e.code {
                    if e.msgSubstr == "" || strings.Contains(strings.ToLower(d.Message), strings.ToLower(e.msgSubstr)) {
                        matches++
                    }
                }
            }
            if e.countSet {
                if matches != e.count {
                    return false, diags
                }
            } else {
                if matches < 1 {
                    return false, diags
                }
            }
        case "warn":
            matches := 0
            for _, d := range diags {
                if d.Level == diag.Warn && d.Code == e.code {
                    if e.msgSubstr == "" || strings.Contains(strings.ToLower(d.Message), strings.ToLower(e.msgSubstr)) {
                        matches++
                    }
                }
            }
            if e.countSet {
                if matches != e.count {
                    return false, diags
                }
            } else {
                if matches < 1 {
                    return false, diags
                }
            }
        case "errors_count":
            if !e.countSet {
                e.countSet = true
                e.count = 1
            }
            if errCount != e.count {
                return false, diags
            }
        case "warnings_count":
            if !e.countSet {
                e.countSet = true
                e.count = 1
            }
            if warnCount != e.count {
                return false, diags
            }
        }
    }
    // No expectations means default pass on no-errors
    if len(c.expects) == 0 {
        return !hasErr, diags
    }
    return true, nil
}

