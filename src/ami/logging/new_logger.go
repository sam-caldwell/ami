package logging

import (
    "io"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"

    stdlogger "github.com/sam-caldwell/ami/src/ami/stdlib/logger"
)

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
            // Resolve pipeline defaults with env and Options overrides.
            // Precedence: Options > Environment > Defaults.
            const (
                defCap  = 256
                defBatch = 1
                // enable time-based flush by default for smoother async writes
                defFlush = 250 * time.Millisecond
            )
            // helpers
            getenvInt := func(key string) (int, bool) {
                if v, ok := os.LookupEnv(key); ok {
                    if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil { return n, true }
                }
                return 0, false
            }
            getenvDur := func(key string) (time.Duration, bool) {
                if v, ok := os.LookupEnv(key); ok {
                    if d, err := time.ParseDuration(strings.TrimSpace(v)); err == nil { return d, true }
                }
                return 0, false
            }
            getenvPolicy := func(key string) (stdlogger.BackpressurePolicy, bool) {
                if v, ok := os.LookupEnv(key); ok {
                    switch strings.ToLower(strings.TrimSpace(v)) {
                    case string(stdlogger.Block):
                        return stdlogger.Block, true
                    case strings.ToLower(string(stdlogger.DropNewest)):
                        return stdlogger.DropNewest, true
                    case strings.ToLower(string(stdlogger.DropOldest)):
                        return stdlogger.DropOldest, true
                    }
                }
                return "", false
            }
            // capacity
            capVal := defCap
            if n, ok := getenvInt("AMI_LOG_PIPE_CAPACITY"); ok { capVal = n }
            if opts.PipelineCapacity > 0 { capVal = opts.PipelineCapacity }
            // batchMax
            batchVal := defBatch
            if n, ok := getenvInt("AMI_LOG_PIPE_BATCH_MAX"); ok { batchVal = n }
            if opts.PipelineBatchMax > 0 { batchVal = opts.PipelineBatchMax }
            // flush interval
            flushVal := defFlush
            if d, ok := getenvDur("AMI_LOG_PIPE_FLUSH_INTERVAL"); ok { flushVal = d }
            if opts.PipelineFlushInterval > 0 { flushVal = opts.PipelineFlushInterval }
            // policy
            polVal := stdlogger.Block
            if p, ok := getenvPolicy("AMI_LOG_PIPE_POLICY"); ok { polVal = p }
            if opts.PipelinePolicy != "" {
                switch strings.ToLower(strings.TrimSpace(opts.PipelinePolicy)) {
                case string(stdlogger.Block):
                    polVal = stdlogger.Block
                case strings.ToLower(string(stdlogger.DropNewest)):
                    polVal = stdlogger.DropNewest
                case strings.ToLower(string(stdlogger.DropOldest)):
                    polVal = stdlogger.DropOldest
                }
            }

            cfg := stdlogger.Config{
                Capacity:           capVal,
                BatchMax:           batchVal,
                FlushInterval:      flushVal,
                Policy:             polVal,
                Sink:               sink,
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

