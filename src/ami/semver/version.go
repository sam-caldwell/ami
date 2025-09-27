package semver

import (
    "fmt"
    "strconv"
    "strings"
)

// Version represents a semantic version (major.minor.patch[-prerelease]).
type Version struct {
    Major int
    Minor int
    Patch int
    Pre   string // optional prerelease tag
}

// ParseVersion parses a semantic version string (with optional leading 'v').
func ParseVersion(s string) (Version, error) {
    s = strings.TrimSpace(s)
    if strings.HasPrefix(s, "v") { s = s[1:] }
    parts := strings.SplitN(s, "-", 2)
    core := parts[0]
    var pre string
    if len(parts) == 2 { pre = parts[1] }
    nums := strings.Split(core, ".")
    if len(nums) != 3 { return Version{}, fmt.Errorf("invalid semver: %s", s) }
    maj, err := strconv.Atoi(nums[0]); if err != nil { return Version{}, err }
    min, err := strconv.Atoi(nums[1]); if err != nil { return Version{}, err }
    pat, err := strconv.Atoi(nums[2]); if err != nil { return Version{}, err }
    return Version{Major: maj, Minor: min, Patch: pat, Pre: pre}, nil
}

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

