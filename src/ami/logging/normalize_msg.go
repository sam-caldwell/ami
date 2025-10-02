package logging

import "strings"

// normalizeMsg converts CRLF to LF to keep outputs deterministic.
func normalizeMsg(s string) string {
    return strings.ReplaceAll(s, "\r\n", "\n")
}

