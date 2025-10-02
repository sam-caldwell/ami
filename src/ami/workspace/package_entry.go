package workspace

// PackageEntry binds a logical key (e.g., "main") to a Package.
type PackageEntry struct {
    Key     string
    Package Package
}

