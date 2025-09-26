package root

import (
    "bufio"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"
    "os/exec"
    "sort"
    "strconv"
    "strings"
    "time"

    ex "github.com/sam-caldwell/ami/src/internal/exit"
    "github.com/sam-caldwell/ami/src/internal/logger"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

// goTestEvent mirrors the subset of fields from `go test -json` we care about
type goTestEvent struct {
    Time    time.Time `json:"Time"`
    Action  string    `json:"Action"` // run|pass|fail|skip|output
    Package string    `json:"Package"`
    Test    string    `json:"Test"`
    Elapsed float64   `json:"Elapsed"`
    Output  string    `json:"Output"`
}

// per-package counters
type ctot struct{ pass, fail, skip int }

func runGoTests(patterns []string) int {
    start := time.Now()

    // Optional fallback: read package parallelism from env when not set via flag
    if testPkgParallel <= 0 {
        if s := os.Getenv("AMI_TEST_PKG_PARALLEL"); s != "" {
            if v, err := strconv.Atoi(s); err == nil && v > 0 {
                testPkgParallel = v
            }
        }
    }

    // Emit run_start in JSON mode with the patterns; resolving packages is optional
    if flagJSON {
        ev := sch.TestRunStart{Schema: "test.v1", Type: "run_start", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Workspace: ".", Packages: patterns,
            Timeout: testTimeout, Parallel: testParallel, PkgParallel: testPkgParallel, FailFast: testFailFast, Run: testRunFilter,
        }
        if err := ev.Validate(); err == nil {
            _ = json.NewEncoder(os.Stdout).Encode(ev)
        }
    } else {
        logger.Info(fmt.Sprintf("test: running %s", strings.Join(patterns, " ")), nil)
    }

    // Invoke `go test -json` once for all patterns
    args := []string{"test", "-json", "-count=1"}
    if testTimeout != "" {
        args = append(args, "-timeout", testTimeout)
    }
    if testParallel > 0 {
        args = append(args, "-parallel", strconv.Itoa(testParallel))
    }
    if testPkgParallel > 0 {
        args = append(args, "-p", strconv.Itoa(testPkgParallel))
    }
    if testFailFast {
        args = append(args, "-failfast")
    }
    if testRunFilter != "" {
        args = append(args, "-run", testRunFilter)
    }
    args = append(args, patterns...)
    cmd := exec.Command("go", args...)
    // We only need stdout (test JSON); capture stderr for diagnostics when human
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        logger.Error(fmt.Sprintf("failed to start go test: %v", err), nil)
        return ex.SystemIOError
    }
    var stderr io.ReadCloser
    if !flagJSON {
        stderr, _ = cmd.StderrPipe()
    }
    if err := cmd.Start(); err != nil {
        logger.Error(fmt.Sprintf("failed to run go test: %v", err), nil)
        return ex.SystemIOError
    }

    // Totals (global and per-package)
    var pass, fail, skip, cases int
    pkgTotals := map[string]*ctot{}
    get := func(pkg string) *ctot {
        if v, ok := pkgTotals[pkg]; ok {
            return v
        }
        v := &ctot{}
        pkgTotals[pkg] = v
        return v
    }

    // For human mode, print simple status lines
    printHuman := func(status, pkg, name string, ms int64) {
        switch status {
        case "pass":
            logger.Info(fmt.Sprintf("test PASS %s %s (%dms)", pkg, name, ms), nil)
        case "fail":
            logger.Error(fmt.Sprintf("test FAIL %s %s (%dms)", pkg, name, ms), nil)
        case "skip":
            logger.Warn(fmt.Sprintf("test SKIP %s %s", pkg, name), nil)
        }
    }

    // Stream JSON from go test
    sc := bufio.NewScanner(stdout)
    sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
    for sc.Scan() {
        line := strings.TrimSpace(sc.Text())
        if line == "" {
            continue
        }
        var ev goTestEvent
        if json.Unmarshal([]byte(line), &ev) != nil {
            // Ignore non-JSON lines in stdout (unlikely)
            continue
        }
        // Only per-test events (Test non-empty) map to our schema
        if ev.Test == "" {
            continue
        }
        switch ev.Action {
        case "run":
            cases++
            if flagJSON {
                s := sch.TestStart{Schema: "test.v1", Type: "test_start", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Package: ev.Package, Name: ev.Test}
                if s.Validate() == nil {
                    _ = json.NewEncoder(os.Stdout).Encode(s)
                }
            } else {
                logger.Info(fmt.Sprintf("test RUN  %s %s", ev.Package, ev.Test), nil)
            }
        case "output":
            if flagJSON {
                o := sch.TestOutput{Schema: "test.v1", Type: "test_output", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Package: ev.Package, Name: ev.Test, Stream: "stdout", Text: ev.Output}
                if o.Validate() == nil {
                    _ = json.NewEncoder(os.Stdout).Encode(o)
                }
            }
        case "pass":
            pass++
            dur := int64(ev.Elapsed * 1000)
            if flagJSON {
                e := sch.TestEnd{Schema: "test.v1", Type: "test_end", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Package: ev.Package, Name: ev.Test, Status: "pass", DurationMs: dur}
                if e.Validate() == nil {
                    _ = json.NewEncoder(os.Stdout).Encode(e)
                }
            } else {
                printHuman("pass", ev.Package, ev.Test, dur)
            }
            get(ev.Package).pass++
        case "fail":
            fail++
            dur := int64(ev.Elapsed * 1000)
            if flagJSON {
                e := sch.TestEnd{Schema: "test.v1", Type: "test_end", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Package: ev.Package, Name: ev.Test, Status: "fail", DurationMs: dur}
                if e.Validate() == nil {
                    _ = json.NewEncoder(os.Stdout).Encode(e)
                }
            } else {
                printHuman("fail", ev.Package, ev.Test, dur)
            }
            get(ev.Package).fail++
        case "skip":
            skip++
            dur := int64(ev.Elapsed * 1000)
            if flagJSON {
                e := sch.TestEnd{Schema: "test.v1", Type: "test_end", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Package: ev.Package, Name: ev.Test, Status: "skip", DurationMs: dur}
                if e.Validate() == nil {
                    _ = json.NewEncoder(os.Stdout).Encode(e)
                }
            } else {
                printHuman("skip", ev.Package, ev.Test, dur)
            }
            get(ev.Package).skip++
        }
    }
    _ = stdout.Close()
    // In human mode, mirror stderr if present (e.g., build failures)
    if !flagJSON && stderr != nil {
        b, _ := io.ReadAll(stderr)
        s := strings.TrimSpace(string(b))
        if s != "" {
            logger.Error(s, nil)
        }
    }
    // Wait for process and decide exit
    err := cmd.Wait()
    // After Go tests, run native AMI tests (discovery under workspace)
    // This step updates pass/fail/skip totals and per-package counts and emits test events.
    runAmiTests(&pass, &fail, &skip, get)

    // Emit run_end
    if flagJSON {
        end := sch.TestRunEnd{Schema: "test.v1", Type: "run_end", Timestamp: sch.FormatTimestamp(time.Now().UTC()), DurationMs: time.Since(start).Milliseconds()}
        end.Totals.Pass = pass
        end.Totals.Fail = fail
        end.Totals.Skip = skip
        end.Totals.Cases = pass + fail + skip
        // Build per-package summaries in deterministic order
        var names []string
        for n := range pkgTotals {
            names = append(names, n)
        }
        sort.Strings(names)
        for _, n := range names {
            ct := pkgTotals[n]
            end.Packages = append(end.Packages, sch.TestPackageSummary{Package: n, Pass: ct.pass, Fail: ct.fail, Skip: ct.skip, Cases: ct.pass + ct.fail + ct.skip})
        }
        if end.Validate() == nil {
            _ = json.NewEncoder(os.Stdout).Encode(end)
        }
    } else {
        // Per-package summaries
        var names []string
        for n := range pkgTotals {
            names = append(names, n)
        }
        sort.Strings(names)
        for _, n := range names {
            ct := pkgTotals[n]
            cases := ct.pass + ct.fail + ct.skip
            logger.Info(fmt.Sprintf("test: package %s — %d pass, %d fail, %d skip, %d cases", n, ct.pass, ct.fail, ct.skip, cases), nil)
        }
        logger.Info(fmt.Sprintf("test: summary: %d pass, %d fail, %d skip", pass, fail, skip), nil)
    }

    if err == nil {
        if fail > 0 {
            return ex.UserError
        }
        return ex.Success
    }
    var ee *exec.ExitError
    if errors.As(err, &ee) {
        // `go test` uses exit 1 for test failures. If we saw fails, map to USER_ERROR.
        if fail > 0 {
            return ex.UserError
        }
        // Otherwise, treat as system I/O error (build failure, test binary error)
        return ex.SystemIOError
    }
    return ex.SystemIOError
}

