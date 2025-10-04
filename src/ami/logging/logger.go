package logging

import (
    "io"
    "time"

    stdlogger "github.com/sam-caldwell/ami/src/ami/runtime/host/logger"
)

// Logger is a simple structured logger with JSON/human formats.
type Logger struct {
    pkg       string
    formatter Formatter
    out       io.Writer
    // alsoWrite, if non-nil, receives a copy of each formatted line (e.g., debug file)
    alsoWrite io.Writer
    // optional pipeline for debug writes
    pipe *stdlogger.Pipeline
    redactKeys      []string
    redactPrefixes  []string
    allowKeys       []string
    denyKeys        []string
}

// Close closes any additional sinks owned by the logger.
func (l *Logger) Close() error {
    if l.pipe != nil { l.pipe.Close() }
    if c, ok := l.alsoWrite.(io.Closer); ok && c != nil { return c.Close() }
    return nil
}

// log emits a log record with given level, message, and optional fields.
func (l *Logger) log(level Level, msg string, fields map[string]any) {
    if fields == nil {
        fields = map[string]any{}
    }
    fields = filterRedactFields(fields, l.allowKeys, l.denyKeys, l.redactKeys, l.redactPrefixes)
    rec := Record{
        Timestamp: time.Now().UTC(),
        Level:     level,
        Message:   msg,
        Package:   l.pkg,
        Fields:    fields,
    }
    line := l.formatter.Format(rec)
    if l.out != nil { _, _ = l.out.Write(line) }
    if l.pipe != nil { _ = l.pipe.Enqueue(line) } else if l.alsoWrite != nil { _, _ = l.alsoWrite.Write(line) }
}

// Trace logs at trace level.
func (l *Logger) Trace(msg string, fields map[string]any) { l.log(LevelTrace, msg, fields) }

// Debug logs at debug level.
func (l *Logger) Debug(msg string, fields map[string]any) { l.log(LevelDebug, msg, fields) }

// Info logs at info level.
func (l *Logger) Info(msg string, fields map[string]any) { l.log(LevelInfo, msg, fields) }

// Warn logs at warn level.
func (l *Logger) Warn(msg string, fields map[string]any) { l.log(LevelWarn, msg, fields) }

// Error logs at error level.
func (l *Logger) Error(msg string, fields map[string]any) { l.log(LevelError, msg, fields) }

// Fatal logs at fatal level.
func (l *Logger) Fatal(msg string, fields map[string]any) { l.log(LevelFatal, msg, fields) }

// PipelineStats returns the current stdlib pipeline counters, if a pipeline is active.
// The boolean indicates whether a pipeline is present (true) or not (false).
func (l *Logger) PipelineStats() (stdlogger.Stats, bool) {
    if l.pipe == nil { return stdlogger.Stats{}, false }
    return l.pipe.Stats(), true
}
