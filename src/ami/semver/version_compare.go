package semver

// Compare returns -1 if a<b, 0 if a==b, 1 if a>b (SemVer precedence; prerelease < release).
func Compare(a, b Version) int {
    if a.Major != b.Major { if a.Major < b.Major { return -1 } ; return 1 }
    if a.Minor != b.Minor { if a.Minor < b.Minor { return -1 } ; return 1 }
    if a.Patch != b.Patch { if a.Patch < b.Patch { return -1 } ; return 1 }
    // Prerelease has lower precedence than no prerelease
    if a.Pre == b.Pre { return 0 }
    if a.Pre == "" { return 1 }
    if b.Pre == "" { return -1 }
    // Simple lexicographic compare for prerelease tags
    if a.Pre < b.Pre { return -1 }
    if a.Pre > b.Pre { return 1 }
    return 0
}

