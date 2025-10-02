package semver

// ValidateVersion returns true when s is a valid semver (optionally prefixed with v).
func ValidateVersion(s string) bool { return reVersion.MatchString(s) }

