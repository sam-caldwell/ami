package workspace

// HighestSatisfying returns the highest version string from the provided list
// that satisfies the given constraint. If includePrerelease is false, versions
// with a pre-release suffix are excluded. The returned bool indicates a match
// was found.
func HighestSatisfying(versions []string, c Constraint, includePrerelease bool) (string, bool) {
    var best string
    var bestV semver
    have := false
    for _, vraw := range versions {
        v, err := parseSemver(trimV(vraw))
        if err != nil { continue }
        if !includePrerelease && v.Pre != "" { continue }
        if !Satisfies(vraw, c) { continue }
        if !have || cmpSemver(v, bestV) > 0 {
            best, bestV, have = vraw, v, true
        }
    }
    return best, have
}

