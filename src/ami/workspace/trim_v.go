package workspace

import "strings"

func trimV(s string) string { return strings.TrimPrefix(strings.TrimSpace(s), "v") }

