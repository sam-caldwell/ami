package workspace

// semver is a minimal semantic version model.
type semver struct {
    Major int
    Minor int
    Patch int
    Pre   string // optional pre-release; ignored for range bounds
}

