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
    out     io.Writer
    err     io.Writer
}

var std = &Logger{out: os.Stdout, err: os.Stderr}

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
    return t.UTC().Format("2006-01-02T15:04:05.000Z07:00")
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
    fmt.Fprintln(l.out, string(b))
}

func (l *Logger) logHuman(w io.Writer, level, msg string) {
    ts := ""
    if l.verbose {
        ts = FormatTimestamp(time.Now()) + " "
    }
    // Prefix every line in multi-line messages
    for _, line := range strings.Split(strings.TrimRight(msg, "\n"), "\n") {
        fmt.Fprintln(w, ts+line)
    }
}

func Info(msg string, data map[string]interface{}) {
    std.mu.Lock(); defer std.mu.Unlock()
    if std.json {
        std.logJSON("info", msg, data)
    } else {
        std.logHuman(std.out, "info", msg)
    }
}

func Warn(msg string, data map[string]interface{}) {
    std.mu.Lock(); defer std.mu.Unlock()
    if std.json {
        std.logJSON("warn", msg, data)
    } else {
        std.logHuman(std.err, "warn", msg)
    }
}

func Error(msg string, data map[string]interface{}) {
    std.mu.Lock(); defer std.mu.Unlock()
    if std.json {
        std.logJSON("error", msg, data)
    } else {
        std.logHuman(std.err, "error", msg)
    }
}

