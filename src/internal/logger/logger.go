package logger

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "strings"
    "sync"
    "time"
)

type Logger struct {
    json    bool
    verbose bool
    color   bool
    mu      sync.Mutex
}

var std = &Logger{}

func Setup(jsonMode, verbose, color bool) {
    std.mu.Lock()
    defer std.mu.Unlock()
    std.json = jsonMode
    // If JSON, force color off
    if jsonMode {
        color = false
    }
    std.verbose = verbose
    std.color = color
}

func FormatTimestamp(t time.Time) string {
    // ISO-8601 UTC with milliseconds and 'Z' suffix
    return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

type record struct {
    Schema    string                 `json:"schema"`
    Timestamp string                 `json:"timestamp"`
    Level     string                 `json:"level"`
    Message   string                 `json:"message"`
    Data      map[string]interface{} `json:"data,omitempty"`
}

func (l *Logger) logJSON(level, msg string, data map[string]interface{}) {
    rec := record{
        Schema:    "diag.v1",
        Timestamp: FormatTimestamp(time.Now()),
        Level:     level,
        Message:   msg,
        Data:      data,
    }
    b, _ := json.Marshal(rec)
    // JSON output always goes to stdout
    fmt.Fprintln(os.Stdout, string(b))
}

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
        if startColor != "" { fmt.Fprintln(w, startColor+ts+line+endColor) } else { fmt.Fprintln(w, ts+line) }
    }
}

func Info(msg string, data map[string]interface{}) {
    std.mu.Lock(); defer std.mu.Unlock()
    if std.json {
        std.logJSON("info", msg, data)
    } else {
        std.logHuman(os.Stdout, "info", msg)
    }
}

func Warn(msg string, data map[string]interface{}) {
    std.mu.Lock(); defer std.mu.Unlock()
    if std.json { std.logJSON("warn", msg, data) } else { std.logHuman(os.Stderr, "warn", msg) }
}

func Error(msg string, data map[string]interface{}) {
    std.mu.Lock(); defer std.mu.Unlock()
    if std.json { std.logJSON("error", msg, data) } else { std.logHuman(os.Stderr, "error", msg) }
}
