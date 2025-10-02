package semver

// ValidateConstraint returns true when the constraint string is valid per ParseConstraint.
func ValidateConstraint(s string) bool {
    _, err := ParseConstraint(s)
    return err == nil
}

