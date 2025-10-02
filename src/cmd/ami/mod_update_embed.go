package main

import "github.com/sam-caldwell/ami/src/ami/workspace"

func embedAudit(r workspace.AuditReport) *modAuditEmbed {
    return &modAuditEmbed{
        MissingInSum:   r.MissingInSum,
        Unsatisfied:    r.Unsatisfied,
        MissingInCache: r.MissingInCache,
        Mismatched:     r.Mismatched,
        ParseErrors:    r.ParseErrors,
        SumFound:       r.SumFound,
    }
}

