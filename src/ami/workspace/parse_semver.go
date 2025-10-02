package workspace

import (
    "fmt"
    "strconv"
    "strings"
)

func parseSemver(s string) (semver, error) {
    // s = MAJOR.MINOR.PATCH[-PRERELEASE]
    parts := strings.SplitN(s, "-", 2)
    nums := strings.Split(parts[0], ".")
    if len(nums) != 3 { return semver{}, fmt.Errorf("invalid semver: %q", s) }
    maj, err := strconv.Atoi(nums[0]); if err != nil { return semver{}, err }
    min, err := strconv.Atoi(nums[1]); if err != nil { return semver{}, err }
    pat, err := strconv.Atoi(nums[2]); if err != nil { return semver{}, err }
    v := semver{Major: maj, Minor: min, Patch: pat}
    if len(parts) == 2 { v.Pre = parts[1] }
    return v, nil
}

