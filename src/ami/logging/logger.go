package logging

import (
    "io"
    "os"
    "path/filepath"
    "time"

    stdlogger "github.com/sam-caldwell/ami/src/ami/stdlib/logger"
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

// New creates a Logger configured by Options.
func New(opts Options) (*Logger, error) {
    l := &Logger{}
    l.pkg = opts.Package

    // choose formatter: JSON always (even when --json is not used)
    // HumanFormatter remains available for non-logger human outputs, but logger emits JSON.
    l.formatter = JSONFormatter{}

    // primary writer
    if opts.Out != nil {
        l.out = opts.Out
    } else {
        l.out = os.Stdout
    }

    // optional debug sink when verbose: write to build/debug/activity.log
    if opts.Verbose {
        debugDir := opts.DebugDir
        if debugDir == "" {
            debugDir = filepath.Join("build", "debug")
        }
        if err := os.MkdirAll(debugDir, 0o755); err == nil {
            // Wire stdlib logger pipeline to file sink for verbose debug logs.
            sink := stdlogger.NewFileSink(filepath.Join(debugDir, "activity.log"), 0o644)
            cfg := stdlogger.Config{
                Capacity:         256,
                BatchMax:         1,             // preserve line-by-line behavior by default
                FlushInterval:    0,              // no timer by default
                Policy:           stdlogger.Block, // preserve current behavior
                Sink:             sink,
                JSONRedactKeys:     opts.RedactKeys,
                JSONRedactPrefixes: opts.RedactPrefixes,
                JSONAllowKeys:      opts.FilterAllowKeys,
                JSONDenyKeys:       opts.FilterDenyKeys,
            }
            p := stdlogger.NewPipeline(cfg)
            if err := p.Start(); err == nil {
                l.pipe = p
            } else {
                // Fallback to direct file append if pipeline fails to start
                if f, ferr := os.OpenFile(filepath.Join(debugDir, "activity.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); ferr == nil {
                    l.alsoWrite = f
                }
            }
        }
    }

    l.redactKeys = opts.RedactKeys
    l.redactPrefixes = opts.RedactPrefixes
    l.allowKeys = opts.FilterAllowKeys
    l.denyKeys = opts.FilterDenyKeys
    return l, nil
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
