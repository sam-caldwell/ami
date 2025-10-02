package semver

// Version represents a semantic version (major.minor.patch[-prerelease]).
type Version struct {
    Major int
    Minor int
    Patch int
    Pre   string // optional prerelease tag
}
 
