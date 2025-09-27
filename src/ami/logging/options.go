package logging

import "io"

// Options configures the Logger behavior.
type Options struct {
    // JSON selects JSON output when true; otherwise human-readable.
    JSON bool
    // Verbose prefixes human lines with ISO-8601 UTC timestamp (ms).
    Verbose bool
    // Color enables ANSI colors in human mode only.
    Color bool
    // Package name to include in records (optional).
    Package string
    // Out is the primary writer for log lines (defaults to stdout).
    Out io.Writer
    // DebugDir is the base directory for debug artifacts when Verbose.
    // Defaults to "build/debug".
    DebugDir string
    // RedactKeys: list of field keys to redact in structured fields.
    RedactKeys []string
    // RedactPrefixes: mask any field whose key starts with one of these prefixes.
    RedactPrefixes []string
    // FilterAllowKeys: if non-empty, include only these keys in fields; others are dropped.
    FilterAllowKeys []string
    // FilterDenyKeys: drop these keys from fields.
    FilterDenyKeys []string
}
