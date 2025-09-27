package logging

import (
    "bytes"
    "strings"
)

// HumanFormatter renders human-readable lines. Colors apply only when Color=true.
type HumanFormatter struct {
    Verbose bool
    Color   bool
}

func (f HumanFormatter) Format(r Record) []byte {
    // Normalize line endings first.
    msg := normalizeMsg(r.Message)
    lines := strings.Split(msg, "\n")

    var buf bytes.Buffer
    var prefix string
    if f.Verbose {
        prefix = iso8601UTCms(r.Timestamp) + " "
    }
    // Level token in uppercase
    levelToken := "[" + strings.ToUpper(string(r.Level)) + "]"
    if f.Color {
        if c := levelColor(r.Level); c != "" {
            levelToken = c + levelToken + ansiReset
        }
    }

    for i, line := range lines {
        if i > 0 {
            buf.WriteByte('\n')
        }
        buf.WriteString(prefix)
        buf.WriteString(levelToken)
        if r.Package != "" {
            buf.WriteString(" ")
            buf.WriteString(r.Package)
            buf.WriteString(":")
        }
        if line != "" {
            buf.WriteString(" ")
            buf.WriteString(line)
        }
    }
    buf.WriteByte('\n')
    return buf.Bytes()
}
