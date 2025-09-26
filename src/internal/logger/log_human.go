package logger

import (
    "fmt"
    "io"
    "os"
    "strings"
    "time"
)

func (l *Logger) logHuman(w io.Writer, level, msg string) {
    ts := ""
    if l.verbose {
        ts = FormatTimestamp(time.Now()) + " "
    }
    // Apply optional color to the full line in human mode
    startColor := ""
    endColor := ""
    if l.color {
        switch level {
        case "error":
            startColor = "\x1b[31m" // red
        case "warn":
            startColor = "\x1b[33m" // yellow
        case "info":
            startColor = "\x1b[32m" // green
        }
        endColor = "\x1b[0m"
    }
    // Prefix every line in multi-line messages
    for _, line := range strings.Split(strings.TrimRight(msg, "\n"), "\n") {
        if startColor != "" {
            fmt.Fprintln(w, startColor+ts+line+endColor)
        } else {
            fmt.Fprintln(w, ts+line)
        }
    }
}

func Info(msg string, data map[string]interface{}) {
    std.mu.Lock()
    defer std.mu.Unlock()
    if std.json {
        std.logJSON("info", msg, data)
    } else {
        std.logHuman(os.Stdout, "info", msg)
    }
}

func Warn(msg string, data map[string]interface{}) {
    std.mu.Lock()
    defer std.mu.Unlock()
    if std.json {
        std.logJSON("warn", msg, data)
    } else {
        std.logHuman(os.Stderr, "warn", msg)
    }
}

func Error(msg string, data map[string]interface{}) {
    std.mu.Lock()
    defer std.mu.Unlock()
    if std.json {
        std.logJSON("error", msg, data)
    } else {
        std.logHuman(os.Stderr, "error", msg)
    }
}

