package semver

import (
    "fmt"
    "strconv"
    "strings"
)

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

