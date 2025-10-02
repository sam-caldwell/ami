package workspace

// AuditReport summarizes dependency status for a workspace against ami.sum and cache.
type AuditReport struct {
    // Requirements lists remote dependency requirements discovered across packages.
    Requirements []Requirement
    // MissingInSum lists requirement names not present in ami.sum.
    MissingInSum []string
    // Unsatisfied lists requirement names present in ami.sum but with no version satisfying the constraint.
    Unsatisfied []string
    // MissingInCache lists name@version entries (that satisfy constraints) that are missing from the cache.
    MissingInCache []string
    // Mismatched lists name@version entries (that satisfy constraints) with sha mismatches vs cache contents.
    Mismatched []string
    // ParseErrors captures import constraint parse errors as strings; requirements with parse errors are skipped.
    ParseErrors []string
    // SumFound indicates whether ami.sum existed and was parsed.
    SumFound bool
}

