package semver

// Bound represents a half-open or closed interval over semantic versions.
// Upper == nil means unbounded above. Inclusivity flags control endpoint inclusion.
type Bound struct {
    Lower          Version
    LowerInclusive bool
    Upper          *Version
    UpperInclusive bool
}
