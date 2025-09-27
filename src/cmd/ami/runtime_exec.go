package main

import (
    "bufio"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/runtime/tester"
)

type runtimeCounts struct{
    total, ok, fail, skip int
}

// runRuntime executes runtime cases with concurrency and optional filters.
func runRuntime(dir string, jsonOut bool, verbose bool, out ioWriter, cases []runtimeCase) runtimeCounts {
    // Filter by name if --run provided
    if pat := currentTestOptions.RunPattern; pat != "" {
        re, _ := regexp.Compile(pat)
        filtered := cases[:0]
        for _, c := range cases { if re.MatchString(c.Name) { filtered = append(filtered, c) } }
        cases = filtered
    }
    counts := runtimeCounts{}
    if len(cases) == 0 { return counts }
    // Verbose log file and manifest append
    var logw *bufio.Writer
    var mfw *bufio.Writer
    if verbose {
        _ = os.MkdirAll(filepath.Join(dir, "build", "test"), 0o755)
        if f, err := os.OpenFile(filepath.Join(dir, "build", "test", "runtime.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644); err == nil {
            defer f.Close(); logw = bufio.NewWriter(f)
            defer logw.Flush()
        }
        if f, err := os.OpenFile(filepath.Join(dir, "build", "test", "test.manifest"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err == nil {
            defer f.Close(); mfw = bufio.NewWriter(f)
            defer mfw.Flush()
        }
    }
    // Emit run_start in JSON mode
    start := time.Now()
    if jsonOut {
        _ = json.NewEncoder(out).Encode(map[string]any{"schema":"test.v1","type":"run_start","timeout_ms": currentTestOptions.TimeoutMs, "parallel": currentTestOptions.Parallel})
    }
    // Worker pool
    par := currentTestOptions.Parallel
    if par <= 0 { par = 1 }
    ch := make(chan runtimeCase)
    type result struct{ c runtimeCase; ok bool; skipped bool; err error; dur time.Duration }
    res := make(chan result)
    // start workers
    for i := 0; i < par; i++ {
        go func() {
            for c := range ch {
                if c.Spec.SkipReason != "" { res <- result{c: c, ok: true, skipped: true, dur: 0}; continue }
                // Validate fixtures
                valid := true
                for _, fx := range c.Spec.Fixtures {
                    p := filepath.Clean(filepath.Join(dir, fx.Path))
                    if !strings.HasPrefix(p, filepath.Clean(dir)+string(os.PathSeparator)) { valid = false; break }
                    if st, err := os.Stat(p); err != nil || st.IsDir() { valid = false; break }
                    if fx.Mode != "ro" && fx.Mode != "rw" { valid = false; break }
                }
                if !valid { res <- result{c: c, ok: false, skipped: false, err: errors.New("invalid fixtures"), dur: 0}; continue }
                // Build context
                ctx := context.Background()
                to := time.Duration(currentTestOptions.TimeoutMs) * time.Millisecond
                if c.Spec.TimeoutMs > 0 { to = time.Duration(c.Spec.TimeoutMs) * time.Millisecond }
                if to > 0 { var cancel context.CancelFunc; ctx, cancel = context.WithTimeout(ctx, to); defer cancel() }
                // Parse input JSON
                var input map[string]any
                if c.Spec.InputJSON != "" { _ = json.Unmarshal([]byte(c.Spec.InputJSON), &input) }
                // Execute
                r, err := tester.Run(ctx, tester.Options{Timeout: to}, input)
                // Compare expectation
                ok := err == nil && c.Spec.ExpectError == ""
                if c.Spec.ExpectError != "" { ok = err != nil }
                // If an output expectation is provided, deep-equal
                if ok && c.Spec.ExpectJSON != "" {
                    var want any
                    var got any = r.Output
                    _ = json.Unmarshal([]byte(c.Spec.ExpectJSON), &want)
                    wb, _ := json.Marshal(want)
                    gb, _ := json.Marshal(got)
                    ok = string(wb) == string(gb)
                }
                res <- result{c: c, ok: ok, skipped: false, err: err, dur: r.Duration}
            }
        }()
    }
    // feed cases
    go func(){ for _, c := range cases { ch <- c }; close(ch) }()
    // collect
    for i := 0; i < len(cases); i++ {
        r := <-res
        counts.total++
        if r.skipped { counts.skip++ } else if r.ok { counts.ok++ } else { counts.fail++ }
        // verbose manifest and logs
        if mfw != nil { _, _ = mfw.WriteString(fmt.Sprintf("runtime:%s %s\n", r.c.File, r.c.Name)) }
        if logw != nil {
            st := "ok"; if r.skipped { st = "skip" } else if !r.ok { st = "fail" }
            _, _ = logw.WriteString(fmt.Sprintf("%s %s %s dur=%dms\n", r.c.File, r.c.Name, st, r.dur.Milliseconds()))
        }
        // emit per-test events
        if jsonOut {
            _ = json.NewEncoder(out).Encode(map[string]any{"schema":"test.v1","type":"test_start","file": r.c.File, "case": r.c.Name})
            ev := map[string]any{"schema":"test.v1","type":"test_end","file": r.c.File, "case": r.c.Name, "ok": r.ok, "skipped": r.skipped, "duration_ms": r.dur.Milliseconds()}
            if r.err != nil { ev["error"] = r.err.Error() }
            _ = json.NewEncoder(out).Encode(ev)
        }
        // human output single-line (optional concise)
        if !jsonOut {
            st := "OK"; if r.skipped { st = "SKIP" } else if !r.ok { st = "FAIL" }
            _, _ = fmt.Fprintf(out, "test: runtime %s %s %s\n", r.c.File, r.c.Name, st)
        }
        if currentTestOptions.Failfast && !r.ok && !r.skipped {
            // drain and break
            break
        }
    }
    if jsonOut {
        _ = json.NewEncoder(out).Encode(map[string]any{"schema":"test.v1","type":"run_end","runtime_tests": counts.total, "runtime_failures": counts.fail, "runtime_skipped": counts.skip, "duration_ms": time.Since(start).Milliseconds()})
    }
    return counts
}

type ioWriter interface{ Write([]byte) (int, error) }
