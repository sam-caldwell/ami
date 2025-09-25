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

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/sem"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    ex "github.com/sam-caldwell/ami/src/internal/exit"
    "github.com/sam-caldwell/ami/src/internal/logger"
    sch "github.com/sam-caldwell/ami/src/schemas"
    "github.com/sam-caldwell/ami/src/ami/runtime/tester"
    "github.com/spf13/cobra"
    "path/filepath"
    "io/fs"
)

var (
    testTimeout   string
    testParallel  int
    testFailFast  bool
    testRunFilter string
    testPkgParallel int
)

// per-package counters
type ctot struct{ pass, fail, skip int }

// goTestEvent mirrors the subset of fields from `go test -json` we care about
type goTestEvent struct {
    Time    time.Time `json:"Time"`
    Action  string    `json:"Action"`  // run|pass|fail|skip|output
    Package string    `json:"Package"`
    Test    string    `json:"Test"`
    Elapsed float64   `json:"Elapsed"`
    Output  string    `json:"Output"`
}

func newTestCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "test [packages...]",
        Short: "Run Go tests (JSON stream supported)",
        Example: `  ami test ./...
  ami --json test ./...`,
        Run: func(cmd *cobra.Command, args []string) {
            // Ensure flag-bound variables are populated (defensive)
            if v, err := cmd.Flags().GetString("timeout"); err == nil { testTimeout = v }
            if v, err := cmd.Flags().GetInt("parallel"); err == nil { testParallel = v }
            if v, err := cmd.Flags().GetInt("pkg-parallel"); err == nil { testPkgParallel = v }
            if v, err := cmd.Flags().GetBool("failfast"); err == nil { testFailFast = v }
            if v, err := cmd.Flags().GetString("run"); err == nil { testRunFilter = v }
            // default to ./...
            if len(args) == 0 { args = []string{"./..."} }
            code := runGoTests(args)
            // Ensure process exit reflects result for callers
            os.Exit(code)
        },
    }
    // Flags: timeout, parallel, failfast, run filter
    cmd.Flags().StringVar(&testTimeout, "timeout", "", "per-package timeout (e.g., 1s, 2m, 10m)")
    cmd.Flags().IntVar(&testParallel, "parallel", 0, "parallelism within package (default GOMAXPROCS)")
    cmd.Flags().IntVar(&testPkgParallel, "pkg-parallel", 0, "number of packages to test in parallel (maps to 'go test -p')")
    cmd.Flags().BoolVar(&testFailFast, "failfast", false, "stop after first test failure")
    cmd.Flags().StringVar(&testRunFilter, "run", "", "run only tests matching regular expression")
    return cmd
}

func runGoTests(patterns []string) int {
    start := time.Now()

    // Optional fallback: read package parallelism from env when not set via flag
    if testPkgParallel <= 0 {
        if s := os.Getenv("AMI_TEST_PKG_PARALLEL"); s != "" {
            if v, err := strconv.Atoi(s); err == nil && v > 0 { testPkgParallel = v }
        }
    }

    // Emit run_start in JSON mode with the patterns; resolving packages is optional
    if flagJSON {
        ev := sch.TestRunStart{Schema: "test.v1", Type: "run_start", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Workspace: ".", Packages: patterns,
            Timeout: testTimeout, Parallel: testParallel, PkgParallel: testPkgParallel, FailFast: testFailFast, Run: testRunFilter,
        }
        if err := ev.Validate(); err == nil { _ = json.NewEncoder(os.Stdout).Encode(ev) }
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
    get := func(pkg string) *ctot { if v, ok := pkgTotals[pkg]; ok { return v }; v := &ctot{}; pkgTotals[pkg] = v; return v }

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
        if line == "" { continue }
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
                if s.Validate() == nil { _ = json.NewEncoder(os.Stdout).Encode(s) }
            } else {
                logger.Info(fmt.Sprintf("test RUN  %s %s", ev.Package, ev.Test), nil)
            }
        case "output":
            if flagJSON {
                o := sch.TestOutput{Schema: "test.v1", Type: "test_output", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Package: ev.Package, Name: ev.Test, Stream: "stdout", Text: ev.Output}
                if o.Validate() == nil { _ = json.NewEncoder(os.Stdout).Encode(o) }
            }
            // suppress verbose output lines in human mode to keep display clean
        case "pass":
            pass++
            dur := int64(ev.Elapsed * 1000)
            if flagJSON {
                e := sch.TestEnd{Schema: "test.v1", Type: "test_end", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Package: ev.Package, Name: ev.Test, Status: "pass", DurationMs: dur}
                if e.Validate() == nil { _ = json.NewEncoder(os.Stdout).Encode(e) }
            } else { printHuman("pass", ev.Package, ev.Test, dur) }
            get(ev.Package).pass++
        case "fail":
            fail++
            dur := int64(ev.Elapsed * 1000)
            if flagJSON {
                e := sch.TestEnd{Schema: "test.v1", Type: "test_end", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Package: ev.Package, Name: ev.Test, Status: "fail", DurationMs: dur}
                if e.Validate() == nil { _ = json.NewEncoder(os.Stdout).Encode(e) }
            } else { printHuman("fail", ev.Package, ev.Test, dur) }
            get(ev.Package).fail++
        case "skip":
            skip++
            dur := int64(ev.Elapsed * 1000)
            if flagJSON {
                e := sch.TestEnd{Schema: "test.v1", Type: "test_end", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Package: ev.Package, Name: ev.Test, Status: "skip", DurationMs: dur}
                if e.Validate() == nil { _ = json.NewEncoder(os.Stdout).Encode(e) }
            } else { printHuman("skip", ev.Package, ev.Test, dur) }
            get(ev.Package).skip++
        }
    }
    _ = stdout.Close()
    // In human mode, mirror stderr if present (e.g., build failures)
    if !flagJSON && stderr != nil {
        b, _ := io.ReadAll(stderr)
        s := strings.TrimSpace(string(b))
        if s != "" { logger.Error(s, nil) }
    }
    // Wait for process and decide exit
    err = cmd.Wait()
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
        for n := range pkgTotals { names = append(names, n) }
        sort.Strings(names)
        for _, n := range names {
            ct := pkgTotals[n]
            end.Packages = append(end.Packages, sch.TestPackageSummary{Package: n, Pass: ct.pass, Fail: ct.fail, Skip: ct.skip, Cases: ct.pass + ct.fail + ct.skip})
        }
        if end.Validate() == nil { _ = json.NewEncoder(os.Stdout).Encode(end) }
    } else {
        // Per-package summaries
        var names []string
        for n := range pkgTotals { names = append(names, n) }
        sort.Strings(names)
        for _, n := range names {
            ct := pkgTotals[n]
            cases := ct.pass + ct.fail + ct.skip
            logger.Info(fmt.Sprintf("test: package %s — %d pass, %d fail, %d skip, %d cases", n, ct.pass, ct.fail, ct.skip, cases), nil)
        }
        logger.Info(fmt.Sprintf("test: summary: %d pass, %d fail, %d skip", pass, fail, skip), nil)
    }

    if err == nil {
        if fail > 0 { return ex.UserError }
        return ex.Success
    }
    var ee *exec.ExitError
    if errors.As(err, &ee) {
        // `go test` uses exit 1 for test failures. If we saw fails, map to USER_ERROR.
        if fail > 0 { return ex.UserError }
        // Otherwise, treat as system I/O error (build failure, test binary error)
        return ex.SystemIOError
    }
    return ex.SystemIOError
}

// ---- AMI native tests (Phase 1: directive-driven assertions) ----

type amiExpect struct {
    kind string // "no_errors" | "no_warnings" | "error" | "warn" | "errors_count" | "warnings_count"
    code string // for error/warn kinds
    countSet bool
    count    int
    msgSubstr string
}

type amiCase struct {
    name   string
    file   string
    pkg    string
    expects []amiExpect
    skipReason string
}

// runAmiTests discovers *_test.ami files under workspace packages and executes
// directive-driven assertions, emitting test.v1 events and updating totals.
func runAmiTests(pass, fail, skip *int, get func(pkg string) *ctot) {
    // Load workspace to discover package roots
    ws, _ := workspace.Load("ami.workspace")
    var roots map[string]string
    if ws == nil {
        roots = map[string]string{"main": "./src"}
    } else {
        roots = parseWorkspacePackages(ws)
        if len(roots) == 0 { roots = map[string]string{"main": "./src"} }
    }
    // Walk each root to find *_test.ami
    files := []string{}
    for _, root := range roots {
        filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
            if err != nil { return nil }
            if d.IsDir() { return nil }
            if strings.HasSuffix(d.Name(), "_test.ami") { files = append(files, path) }
            return nil
        })
    }
    sort.Strings(files)
    for _, file := range files {
        b, err := os.ReadFile(file)
        if err != nil { continue }
        src := string(b)
        p := parser.New(src)
        f := p.ParseFile()
        // derive package name for reporting
        pkgName := f.Package
        if pkgName == "" { pkgName = uPackageFromPath(file, roots) }
        if pkgName == "" { pkgName = "main" }
        cases := deriveAmiCases(file, pkgName, f)
        // Derive runtime cases
        rcases := deriveAmiRuntimeCases(file, pkgName, f)
        // Evaluate diagnostics (parser + sem)
        diags := append([]diag.Diagnostic{}, p.Errors()...)
        semres := sem.AnalyzeFile(f)
        diags = append(diags, semres.Diagnostics...)
        // For each case, emit start/end and update totals
        for _, c := range cases {
            startTs := time.Now().UTC()
            if flagJSON {
                ev := sch.TestStart{Schema: "test.v1", Type: "test_start", Timestamp: sch.FormatTimestamp(startTs), Package: c.pkg, Name: c.name}
                if ev.Validate() == nil { _ = json.NewEncoder(os.Stdout).Encode(ev) }
            } else {
                logger.Info(fmt.Sprintf("test RUN  %s %s", c.pkg, c.name), nil)
            }
            if c.skipReason != "" {
                if flagJSON {
                    e := sch.TestEnd{Schema: "test.v1", Type: "test_end", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Package: c.pkg, Name: c.name, Status: "skip", DurationMs: time.Since(startTs).Milliseconds()}
                    if e.Validate() == nil { _ = json.NewEncoder(os.Stdout).Encode(e) }
                } else { logger.Warn(fmt.Sprintf("test SKIP %s %s", c.pkg, c.name), nil) }
                *skip++
                get(c.pkg).skip++
                continue
            }
            ok, failDiags := evalAmiExpectations(c, diags)
            if ok {
                if flagJSON {
                    e := sch.TestEnd{Schema: "test.v1", Type: "test_end", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Package: c.pkg, Name: c.name, Status: "pass", DurationMs: time.Since(startTs).Milliseconds()}
                    if e.Validate() == nil { _ = json.NewEncoder(os.Stdout).Encode(e) }
                } else { logger.Info(fmt.Sprintf("test PASS %s %s", c.pkg, c.name), nil) }
                *pass++
                get(c.pkg).pass++
            } else {
                if flagJSON {
                    // Attach DIAG records into TestEnd.Diagnostics
                    var dlist []sch.DiagV1
                    for _, d := range failDiags { dlist = append(dlist, d.ToSchema()) }
                    e := sch.TestEnd{Schema: "test.v1", Type: "test_end", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Package: c.pkg, Name: c.name, Status: "fail", DurationMs: time.Since(startTs).Milliseconds(), Diagnostics: dlist}
                    if e.Validate() == nil { _ = json.NewEncoder(os.Stdout).Encode(e) }
                } else { logger.Error(fmt.Sprintf("test FAIL %s %s", c.pkg, c.name), nil) }
                *fail++
                get(c.pkg).fail++
            }
        }

        // Execute runtime cases via deterministic harness (Phase 2 scaffold)
        if len(rcases) > 0 {
            // Group by pipeline for execution
            byPipe := map[string][]tester.Case{}
            // Track order and start times for per-case duration
            order := []string{}
            startTimes := map[string]time.Time{}
            key := func(pkg, name string) string { return pkg+"::"+name }
            for _, rc := range rcases {
                // Emit start event now to preserve order with static cases
                st := time.Now().UTC()
                if flagJSON {
                    ev := sch.TestStart{Schema: "test.v1", Type: "test_start", Timestamp: sch.FormatTimestamp(st), Package: rc.pkg, Name: rc.name}
                    if ev.Validate() == nil { _ = json.NewEncoder(os.Stdout).Encode(ev) }
                } else {
                    logger.Info(fmt.Sprintf("test RUN  %s %s", rc.pkg, rc.name), nil)
                }
                order = append(order, key(rc.pkg, rc.name))
                startTimes[key(rc.pkg, rc.name)] = st
                byPipe[rc.pipeline] = append(byPipe[rc.pipeline], tester.Case{
                    Name:        rc.name,
                    Pipeline:    rc.pipeline,
                    InputJSON:   rc.inputJSON,
                    ExpectJSON:  rc.expectJSON,
                    ExpectError: rc.expectError,
                    TimeoutMs:   rc.timeoutMs,
                })
            }
            r := tester.New()
            // Deterministic iteration over pipelines
            var pipes []string
            for p := range byPipe { pipes = append(pipes, p) }
            sort.Strings(pipes)
            // Map results by case name for lookup
            results := map[string]tester.Result{}
            for _, p := range pipes {
                out, _ := r.Execute(p, byPipe[p])
                for i, oc := range byPipe[p] {
                    if i < len(out) {
                        results[key(pkgName, oc.Name)] = out[i]
                    } else {
                        results[key(pkgName, oc.Name)] = tester.Result{Name: oc.Name, Status: "skip", Error: "runtime unavailable"}
                    }
                }
            }
            // Emit end events using original order
            for _, k := range order {
                // k encodes pkg::name; split
                parts := strings.SplitN(k, "::", 2)
                pkg := parts[0]
                name := parts[1]
                st := startTimes[k]
                res, ok := results[k]
                if !ok { res = tester.Result{Name: name, Status: "skip", Error: "runtime unavailable"} }
                status := res.Status
                // Default duration based on wall time when harness does not provide one
                dur := time.Since(st).Milliseconds()
                switch status {
                case "pass":
                    *pass++
                    get(pkg).pass++
                case "fail":
                    *fail++
                    get(pkg).fail++
                default:
                    status = "skip"
                    *skip++
                    get(pkg).skip++
                }
                if flagJSON {
                    e := sch.TestEnd{Schema: "test.v1", Type: "test_end", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Package: pkg, Name: name, Status: status, DurationMs: dur}
                    if e.Validate() == nil { _ = json.NewEncoder(os.Stdout).Encode(e) }
                } else {
                    switch status {
                    case "pass":
                        logger.Info(fmt.Sprintf("test PASS %s %s (%dms)", pkg, name, dur), nil)
                    case "fail":
                        logger.Error(fmt.Sprintf("test FAIL %s %s (%dms)", pkg, name, dur), nil)
                    default:
                        logger.Warn(fmt.Sprintf("test SKIP %s %s", pkg, name), nil)
                    }
                }
            }
        }
    }
}

// deriveAmiCases builds test cases from top-level #pragma directives in the file.
// Pragmas supported:
//  - test:case <name>
//  - test:expect_no_errors
//  - test:expect_error <CODE>
//  - test:expect_warn <CODE>
//  - test:skip <reason>
// If no cases/pragmas present, a default case named after the file asserts no errors.
func deriveAmiCases(file, pkg string, f *astpkg.File) []amiCase {
    var cases []amiCase
    cur := amiCase{}
    flush := func() {
        if cur.name == "" { return }
        cases = append(cases, cur)
        cur = amiCase{}
    }
    for _, d := range f.Directives {
        switch d.Name {
        case "test:case":
            if cur.name != "" { flush() }
            name := strings.TrimSpace(d.Payload)
            if strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"") { name = strings.Trim(name, "\"") }
            cur = amiCase{name: name, file: file, pkg: pkg}
        case "test:expect_no_errors":
            if cur.name == "" { cur = amiCase{name: filepath.Base(file), file: file, pkg: pkg} }
            cur.expects = append(cur.expects, amiExpect{kind: "no_errors"})
        case "test:expect_error":
            if cur.name == "" { cur = amiCase{name: filepath.Base(file), file: file, pkg: pkg} }
            cur.expects = append(cur.expects, parseExpectWithParams("error", d.Payload))
        case "test:expect_warn":
            if cur.name == "" { cur = amiCase{name: filepath.Base(file), file: file, pkg: pkg} }
            cur.expects = append(cur.expects, parseExpectWithParams("warn", d.Payload))
        case "test:expect_errors":
            if cur.name == "" { cur = amiCase{name: filepath.Base(file), file: file, pkg: pkg} }
            e := parseExpectWithParams("errors_count", d.Payload)
            if !e.countSet { e.countSet = true; e.count = 1 }
            cur.expects = append(cur.expects, e)
        case "test:expect_warnings":
            if cur.name == "" { cur = amiCase{name: filepath.Base(file), file: file, pkg: pkg} }
            e := parseExpectWithParams("warnings_count", d.Payload)
            if !e.countSet { e.countSet = true; e.count = 1 }
            cur.expects = append(cur.expects, e)
        case "test:expect_no_warnings":
            if cur.name == "" { cur = amiCase{name: filepath.Base(file), file: file, pkg: pkg} }
            cur.expects = append(cur.expects, amiExpect{kind: "no_warnings"})
        case "test:skip":
            if cur.name == "" { cur = amiCase{name: filepath.Base(file), file: file, pkg: pkg} }
            cur.skipReason = strings.TrimSpace(d.Payload)
        }
    }
    if cur.name != "" { flush() }
    if len(cases) == 0 {
        // default case: no-errors expectation
        cases = append(cases, amiCase{name: filepath.Base(file), file: file, pkg: pkg, expects: []amiExpect{{kind: "no_errors"}}})
    }
    return cases
}

func evalAmiExpectations(c amiCase, diags []diag.Diagnostic) (bool, []diag.Diagnostic) {
    // Build helper maps
    hasErr := false
    errCount := 0
    warnCount := 0
    for _, d := range diags {
        switch d.Level {
        case diag.Error:
            hasErr = true; errCount++
        case diag.Warn:
            warnCount++
        }
    }
    // All expectations must pass
    for _, e := range c.expects {
        switch e.kind {
        case "no_errors":
            if hasErr {
                // collect errors as failure diagnostics
                var out []diag.Diagnostic
                for _, d := range diags { if d.Level == diag.Error { out = append(out, d) } }
                return false, out
            }
        case "no_warnings":
            if warnCount > 0 {
                var out []diag.Diagnostic
                for _, d := range diags { if d.Level == diag.Warn { out = append(out, d) } }
                return false, out
            }
        case "error":
            matches := 0
            for _, d := range diags {
                if d.Level == diag.Error && d.Code == e.code {
                    if e.msgSubstr == "" || strings.Contains(strings.ToLower(d.Message), strings.ToLower(e.msgSubstr)) {
                        matches++
                    }
                }
            }
            if e.countSet {
                if matches != e.count { return false, diags }
            } else {
                if matches < 1 { return false, diags }
            }
        case "warn":
            matches := 0
            for _, d := range diags {
                if d.Level == diag.Warn && d.Code == e.code {
                    if e.msgSubstr == "" || strings.Contains(strings.ToLower(d.Message), strings.ToLower(e.msgSubstr)) {
                        matches++
                    }
                }
            }
            if e.countSet {
                if matches != e.count { return false, diags }
            } else {
                if matches < 1 { return false, diags }
            }
        case "errors_count":
            if !e.countSet { e.countSet = true; e.count = 1 }
            if errCount != e.count { return false, diags }
        case "warnings_count":
            if !e.countSet { e.countSet = true; e.count = 1 }
            if warnCount != e.count { return false, diags }
        }
    }
    // No expectations means default pass on no-errors
    if len(c.expects) == 0 { return !hasErr, diags }
    return true, nil
}

// parseExpectWithParams parses payload of form:
//   CODE [count=N] [msg~="substr"]
// or for aggregate counts: [count=N]
func parseExpectWithParams(kind, payload string) amiExpect {
    e := amiExpect{kind: kind}
    fields := strings.Fields(payload)
    // The first field may be a code for error/warn kinds
    if (kind == "error" || kind == "warn") && len(fields) > 0 && !strings.Contains(fields[0], "=") && !strings.Contains(fields[0], "~") {
        e.code = fields[0]
        fields = fields[1:]
    }
    for _, f := range fields {
        if strings.HasPrefix(f, "count=") {
            n := strings.TrimPrefix(f, "count=")
            if i, err := strconv.Atoi(n); err == nil && i >= 0 { e.countSet = true; e.count = i }
        } else if strings.HasPrefix(f, "msg~=") {
            m := strings.TrimPrefix(f, "msg~=")
            m = strings.Trim(m, "\"'")
            e.msgSubstr = m
        }
    }
    return e
}

// ---- AMI runtime tests (Phase 2: deterministic harness integration) ----

type amiRuntimeCase struct {
    name        string
    file        string
    pkg         string
    pipeline    string
    inputJSON   string
    expectJSON  string
    expectError string
    timeoutMs   int
}

// deriveAmiRuntimeCases builds runtime cases from `#pragma test:runtime ...` directives.
// It associates the most recent `#pragma test:case` as the case name; if absent, the
// file basename is used. Each `test:runtime` directive flushes a case.
func deriveAmiRuntimeCases(file, pkg string, f *astpkg.File) []amiRuntimeCase {
    var cases []amiRuntimeCase
    curName := ""
    for _, d := range f.Directives {
        switch d.Name {
        case "test:case":
            name := strings.TrimSpace(d.Payload)
            if strings.HasPrefix(name, "\"") && strings.HasSuffix(name, "\"") { name = strings.Trim(name, "\"") }
            curName = name
        case "test:runtime":
            name := curName
            if name == "" { name = filepath.Base(file) }
            rc := amiRuntimeCase{name: name, file: file, pkg: pkg}
            // Parse payload key=value pairs
            kv := parseRuntimePayload(d.Payload)
            rc.pipeline = kv["pipeline"]
            rc.inputJSON = kv["input"]
            rc.expectJSON = kv["expect_output"]
            rc.expectError = kv["expect_error"]
            if t := strings.TrimSpace(kv["timeout"]); t != "" {
                if n, err := strconv.Atoi(t); err == nil && n >= 0 { rc.timeoutMs = n }
            }
            cases = append(cases, rc)
        }
    }
    return cases
}

// parseRuntimePayload parses key=value pairs where value may be quoted or a JSON object/array.
// Supported keys: pipeline, input, expect_output, expect_error, timeout
func parseRuntimePayload(s string) map[string]string {
    out := map[string]string{}
    i := 0
    // helper to skip spaces
    skip := func() { for i < len(s) && s[i] == ' ' { i++ } }
    for i < len(s) {
        skip()
        if i >= len(s) { break }
        // parse key
        ks := i
        for i < len(s) && s[i] != '=' && s[i] != ' ' { i++ }
        if i >= len(s) || s[i] != '=' {
            // skip to next space if malformed
            for i < len(s) && s[i] != ' ' { i++ }
            continue
        }
        key := strings.TrimSpace(s[ks:i])
        i++ // skip '='
        skip()
        if i >= len(s) { out[key] = ""; break }
        // parse value
        var val string
        switch s[i] {
        case '\'', '"':
            quote := s[i]
            i++
            vs := i
            for i < len(s) && s[i] != quote { i++ }
            if i <= len(s) { val = s[vs:i] }
            if i < len(s) { i++ }
        case '{':
            // read balanced braces
            depth := 0
            vs := i
            for i < len(s) {
                if s[i] == '{' { depth++ }
                if s[i] == '}' { depth--; if depth == 0 { i++; break } }
                i++
            }
            val = s[vs:i]
        case '[':
            depth := 0
            vs := i
            for i < len(s) {
                if s[i] == '[' { depth++ }
                if s[i] == ']' { depth--; if depth == 0 { i++; break } }
                i++
            }
            val = s[vs:i]
        default:
            vs := i
            for i < len(s) && s[i] != ' ' { i++ }
            val = s[vs:i]
        }
        out[strings.ToLower(key)] = strings.TrimSpace(val)
        skip()
    }
    return out
}
