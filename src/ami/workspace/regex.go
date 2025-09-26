package workspace

import "regexp"

var osArchRe = regexp.MustCompile(`^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$`)
var semverRe = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

// SemverRe exposes the internal SemVer regex for other packages (read‑only).
func SemverRe() *regexp.Regexp { return semverRe }

