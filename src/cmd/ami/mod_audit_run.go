package main

import (
    "encoding/json"
    "fmt"
    "io"
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/exit"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

type modAuditResult struct {
    Requirements    []workspace.Requirement `json:"requirements"`
    MissingInSum    []string                `json:"missingInSum"`
    Unsatisfied     []string                `json:"unsatisfied"`
    MissingInCache  []string                `json:"missingInCache"`
    Mismatched      []string                `json:"mismatched"`
    ParseErrors     []string                `json:"parseErrors"`
    SumFound        bool                    `json:"sumFound"`
    Timestamp       string                  `json:"timestamp"`
}

func runModAudit(out io.Writer, dir string, jsonOut bool) error {
    rep, err := workspace.AuditDependencies(dir)
    if err != nil {
        if jsonOut {
            _ = json.NewEncoder(out).Encode(map[string]any{"error": err.Error()})
        }
        return exit.New(exit.User, "audit failed: %v", err)
    }
    res := modAuditResult{
        Requirements:   rep.Requirements,
        MissingInSum:    rep.MissingInSum,
        Unsatisfied:     rep.Unsatisfied,
        MissingInCache:  rep.MissingInCache,
        Mismatched:      rep.Mismatched,
        ParseErrors:     rep.ParseErrors,
        SumFound:        rep.SumFound,
        Timestamp:       time.Now().UTC().Format(time.RFC3339Nano),
    }
    if jsonOut {
        return json.NewEncoder(out).Encode(res)
    }
    // Human summary
    if len(res.ParseErrors) > 0 {
        _, _ = fmt.Fprintf(out, "parse errors: %d\n", len(res.ParseErrors))
    }
    _, _ = fmt.Fprintf(out, "requirements: %d\n", len(res.Requirements))
    if !res.SumFound {
        _, _ = fmt.Fprintln(out, "ami.sum: not found")
    }
    if len(res.MissingInSum) > 0 {
        _, _ = fmt.Fprintf(out, "missing in sum: %s\n", strings.Join(res.MissingInSum, ", "))
    }
    if len(res.Unsatisfied) > 0 {
        _, _ = fmt.Fprintf(out, "unsatisfied: %s\n", strings.Join(res.Unsatisfied, ", "))
    }
    if len(res.MissingInCache) > 0 {
        _, _ = fmt.Fprintf(out, "missing in cache: %s\n", strings.Join(res.MissingInCache, ", "))
    }
    if len(res.Mismatched) > 0 {
        _, _ = fmt.Fprintf(out, "mismatched: %s\n", strings.Join(res.Mismatched, ", "))
    }
    if len(res.MissingInSum)+len(res.Unsatisfied)+len(res.MissingInCache)+len(res.Mismatched)+len(res.ParseErrors) == 0 {
        _, _ = fmt.Fprintln(out, "ok: all requirements satisfied and present in cache")
    }
    return nil
}

