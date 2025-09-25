package parser

import (
	"path"
	"regexp"
	"strings"
)

var pkgIdentRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
var importSegRe = regexp.MustCompile(`^[A-Za-z0-9._~-]+$`)
var versionRe = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

// ValidatePackageIdent enforces Chapter 2.1 identifier rules for package names.
func ValidatePackageIdent(id string) bool {
	return pkgIdentRe.MatchString(id)
}

// ValidateVersion checks that a version matches SemVer (optional leading 'v').
// Examples: 0.0.1, v1.2.3, 1.2.3-rc.1, 1.2.3+meta
func ValidateVersion(v string) bool { return versionRe.MatchString(v) }

// ValidateImportPath validates AMI import paths:
// - relative imports allowed with leading ./
// - absolute paths are rejected
// - no empty segments, no '.' or '..' segments
// - segments contain [A-Za-z0-9._~-]
func ValidateImportPath(p string) bool {
	if p == "" {
		return false
	}
	if strings.HasPrefix(p, "/") {
		return false
	}
	// normalize single leading ./ if present for checks
	if strings.HasPrefix(p, "./") {
		p = strings.TrimPrefix(p, "./")
	}
	if p == "" {
		return false
	}
	segs := strings.Split(p, "/")
	for _, seg := range segs {
		if seg == "" || seg == "." || seg == ".." {
			return false
		}
		if !importSegRe.MatchString(seg) {
			return false
		}
	}
	// Prevent path traversal like a/b/../../c
	if strings.Contains(p, "../") || strings.HasSuffix(p, "/..") {
		return false
	}
	// Clean should not produce a path that changes semantics (no leading ../)
	if strings.HasPrefix(path.Clean(p), "../") {
		return false
	}
	return true
}

// ValidateImportConstraint accepts minimal form used in source files: ">= vX.Y.Z" with optional spaces.
func ValidateImportConstraint(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	// Only accept >= for now
	if strings.HasPrefix(s, ">=") {
		v := strings.TrimSpace(strings.TrimPrefix(s, ">="))
		return versionRe.MatchString(v)
	}
	return false
}
