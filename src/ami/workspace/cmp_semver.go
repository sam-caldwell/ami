package workspace

func cmpSemver(a, b semver) int {
    if a.Major != b.Major { if a.Major < b.Major { return -1 } ; return 1 }
    if a.Minor != b.Minor { if a.Minor < b.Minor { return -1 } ; return 1 }
    if a.Patch != b.Patch { if a.Patch < b.Patch { return -1 } ; return 1 }
    // Pre-release: treat empty (release) > any pre-release
    if a.Pre == b.Pre { return 0 }
    if a.Pre == "" { return 1 }
    if b.Pre == "" { return -1 }
    if a.Pre < b.Pre { return -1 }
    if a.Pre > b.Pre { return 1 }
    return 0
}

