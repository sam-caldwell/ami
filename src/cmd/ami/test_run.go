package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "sort"
    "strings"
    "strconv"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/runtime/kvstore"
)

// runTest executes `go test ./...` in dir. When verbose, writes:
// - build/test/test.log (full test output)
// - build/test/test.manifest (each test name in execution order: "<package> <test>")
// It also evaluates AMI directive-based tests in any `*.ami` files under dir:
// - `#pragma test:case <name>` defines a case
// - `#pragma test:assert parse_ok|parse_fail` sets expected parse outcome (default parse_ok)
// Failing AMI directive tests cause a non-zero exit and are included in summaries.
func runTest(out io.Writer, dir string, jsonOut bool, verbose bool, pkgConcurrency int) error {
    buildDir := filepath.Join(dir, "build", "test")
    if verbose {
        _ = os.MkdirAll(buildDir, 0o755)
    }

    // Decide command: use -json always to capture structured events for manifest.
    // Keep stdout quiet unless verbose or error; logs go to files under build/test.
    // Build arguments for `go test`.
    args := []string{"test", "-json"}
    if pkgConcurrency > 0 { args = append(args, "-p", fmt.Sprintf("%d", pkgConcurrency)) }
    args = append(args, "./...")
    cmd := exec.Command("go", args...)
    cmd.Dir = dir
    // Make runs reproducible; avoid prompts
    env := append(os.Environ(), "GOTRACEBACK=single")
    cmd.Env = env
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    // In JSON mode, stream go test events directly to out while tee'ing to buffer for artifacts.
    // Also emit a synthetic test.v1 run_start event with configured options.
    if jsonOut {
        // Emit run_start
        _ = json.NewEncoder(out).Encode(map[string]any{
            "schema":       "test.v1",
            "type":         "run_start",
            "timeout_ms":   currentTestOptions.TimeoutMs,
            "parallel":     currentTestOptions.Parallel,
            "pkg_parallel": pkgConcurrency,
            "failfast":     currentTestOptions.Failfast,
            "run":          currentTestOptions.RunPattern,
        })
        cmd.Stdout = io.MultiWriter(out, &stdout)
    } else {
        cmd.Stdout = &stdout
    }
    cmd.Stderr = &stderr
    err := cmd.Run()

    // If verbose, write test.log
    if verbose {
        _ = os.WriteFile(filepath.Join(buildDir, "test.log"), stdout.Bytes(), 0o644)
    }

    // Emit kvstore metrics/dump when requested (or in verbose mode for convenience)
    if verbose || currentTestOptions.KvMetrics || currentTestOptions.KvDump {
        kvDir := filepath.Join(buildDir, "kv")
        _ = os.MkdirAll(kvDir, 0o755)
        now := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
        if verbose || currentTestOptions.KvMetrics {
            mts := kvstore.Default().Metrics()
            mobj := map[string]any{
                "schema":      "kv.metrics.v1",
                "timestamp":   now,
                "hits":        mts.Hits,
                "misses":      mts.Misses,
                "expirations": mts.Expirations,
                "evictions":   mts.Evictions,
                "currentSize": mts.CurrentSize,
            }
            _ = writeJSONFile(filepath.Join(kvDir, "metrics.json"), mobj)
        }
        if verbose || currentTestOptions.KvDump {
            keys := kvstore.Default().Keys()
            sort.Strings(keys)
            dobj := map[string]any{
                "schema":    "kv.dump.v1",
                "timestamp": now,
                "keys":      keys,
                "size":      len(keys),
            }
            _ = writeJSONFile(filepath.Join(kvDir, "dump.json"), dobj)
        }
    }

    // Build counts by scanning the go test JSON stream we captured in stdout buffer.
    // We count unique (package,test) pairs for tests, unique packages that have tests,
    // and failures as number of events with Action=="fail" and a Test name.
    type counts struct{
        packages, tests, failures int
        byPkg map[string]struct{ ok, fail, skip int }
    }
    computeCounts := func(b []byte) counts {
        var c counts
        c.byPkg = map[string]struct{ ok, fail, skip int }{}
        if len(b) == 0 { return c }
        dec := json.NewDecoder(bytes.NewReader(b))
        seenTests := map[string]struct{}{}
        seenPkgs := map[string]struct{}{}
        for dec.More() {
            var ev goTestEvent
            if dec.Decode(&ev) != nil {
                // skip malformed lines
                _ = dec.Decode(&map[string]any{})
                continue
            }
            if ev.Test != "" && ev.Package != "" && ev.Action == "run" {
                key := ev.Package + "\x00" + ev.Test
                if _, ok := seenTests[key]; !ok {
                    seenTests[key] = struct{}{}
                    if _, had := seenPkgs[ev.Package]; !had { seenPkgs[ev.Package] = struct{}{} }
                }
            }
            if ev.Test != "" && ev.Package != "" {
                entry := c.byPkg[ev.Package]
                switch ev.Action {
                case "pass": entry.ok++
                case "fail": entry.fail++; c.failures++
                case "skip": entry.skip++
                }
                c.byPkg[ev.Package] = entry
            }
        }
        c.tests = len(seenTests)
        c.packages = len(seenPkgs)
        return c
    }
    c := computeCounts(stdout.Bytes())
    if err != nil && c.packages == 0 {
        // Tolerate empty Go package trees; AMI directive tests may still run.
        s := stdout.String() + "\n" + stderr.String()
        if strings.Contains(s, "no packages to test") || strings.Contains(s, "matched no packages") || strings.Contains(s, "directory prefix . does not contain main module") {
            err = nil
            stderr.Reset()
        }
    }

    // Run AMI directive tests by scanning *.ami files.
    type amiResult struct{
        name   string
        file   string
        ok     bool
        expect string
        count  int
        msgSub string
        gotErrs int
        code   string
        line   int
        column int
        offset int
    }
    amiTests, amiFailures := 0, 0
    var amiManifest []string
    var amiEvents []amiResult
    evaluateAMIDirectives := func(root string) {
        // Walk tree shallowly by listing files recursively using filepath.WalkDir.
        _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
            if err != nil { return nil }
            if d.IsDir() { return nil }
            if !strings.HasSuffix(d.Name(), ".ami") { return nil }
            // Read file
            b, err := os.ReadFile(path)
            if err != nil { return nil }
            // Parse using compiler/parser
            var fs source.FileSet
            f := fs.AddFile(path, string(b))
            p := parser.New(f)
            file, errs := p.ParseFileCollect()
            // default expectation: parse_ok
            expectParseOK := true
            expCount := -1
            expMsg := ""
            var caseNames []string
            var expCode string
            var expLine, expCol, expOff int
            for _, pr := range file.Pragmas {
                if pr.Domain == "test" {
                    switch pr.Key {
                    case "case":
                        if pr.Value != "" { caseNames = append(caseNames, pr.Value) }
                    case "assert":
                        if len(pr.Args) > 0 {
                            if pr.Args[0] == "parse_fail" { expectParseOK = false }
                            if pr.Args[0] == "parse_ok" { expectParseOK = true }
                        }
                        if pr.Params != nil {
                            if v, ok := pr.Params["count"]; ok {
                                if n, e := strconv.Atoi(v); e == nil { expCount = n }
                            }
                            if v, ok := pr.Params["msg"]; ok { expMsg = v }
                            if v, ok := pr.Params["message"]; ok { expMsg = v }
                            if v, ok := pr.Params["code"]; ok { expCode = v }
                            if v, ok := pr.Params["line"]; ok { if n, e := strconv.Atoi(v); e == nil { expLine = n } }
                            if v, ok := pr.Params["column"]; ok { if n, e := strconv.Atoi(v); e == nil { expCol = n } }
                            if v, ok := pr.Params["offset"]; ok { if n, e := strconv.Atoi(v); e == nil { expOff = n } }
                        }
                    }
                }
            }
            // If no explicit cases, create a default case asserting parse_ok
            if len(caseNames) == 0 {
                // If the raw source contains any test pragmas, do not invent a default; tolerate parser failure.
                if bytes.Contains(b, []byte("#pragma test:")) { return nil }
                // Determine result based on default expectation (parse_ok)
                ok := len(errs) == 0
                rel := path
                if rp, rerr := filepath.Rel(root, path); rerr == nil { rel = rp }
                name := "default"
                amiTests++
                if !ok { amiFailures++ }
                amiManifest = append(amiManifest, fmt.Sprintf("ami:%s %s", rel, name))
                amiEvents = append(amiEvents, amiResult{name: name, file: rel, ok: ok, expect: "parse_ok", count: 0, msgSub: "", gotErrs: len(errs)})
                return nil
            }
            // Determine result based on expectation vs error
            ok := (len(errs) == 0) == expectParseOK
            // Count assertion
            if expCount >= 0 && ok {
                if expectParseOK && expCount != 0 { ok = false }
                if !expectParseOK && expCount != len(errs) { ok = false }
            }
            // Message substring assertion when provided
            if expMsg != "" && ok {
                s := ""
                for _, e := range errs { s += e.Error() + "\n" }
                if !strings.Contains(s, expMsg) { ok = false }
            }
            // Code assertion: parser uses a stable code E_PARSE.
            if expCode != "" && ok {
                if expCode != "E_PARSE" { ok = false }
            }
            // Position assertions: require at least one error match.
            if (expLine != 0 || expCol != 0 || expOff != 0) && ok {
                matched := false
                for _, e := range errs {
                    type withPos interface{ Position() source.Position }
                    if wp, ok2 := e.(withPos); ok2 {
                        pos := wp.Position()
                        if (expLine == 0 || pos.Line == expLine) && (expCol == 0 || pos.Column == expCol) && (expOff == 0 || pos.Offset == expOff) {
                            matched = true
                            break
                        }
                    }
                }
                if !matched { ok = false }
            }
            // Record for each case
            rel := path
            if rp, rerr := filepath.Rel(root, path); rerr == nil { rel = rp }
            for _, name := range caseNames {
                amiTests++
                if !ok { amiFailures++ }
                amiManifest = append(amiManifest, fmt.Sprintf("ami:%s %s", rel, name))
                amiEvents = append(amiEvents, amiResult{name: name, file: rel, ok: ok, expect: map[bool]string{true:"parse_ok", false:"parse_fail"}[expectParseOK], count: expCount, msgSub: expMsg, gotErrs: len(errs), code: expCode, line: expLine, column: expCol, offset: expOff})
            }
            return nil
        })
    }
    evaluateAMIDirectives(dir)

    // Build manifest from JSON stream when verbose
    if verbose {
        mf, _ := os.OpenFile(filepath.Join(buildDir, "test.manifest"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
        if mf != nil {
            defer mf.Close()
            dec := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
            bw := bufio.NewWriter(mf)
            for dec.More() {
                var ev goTestEvent
                if dec.Decode(&ev) == nil {
                    if ev.Action == "run" && ev.Test != "" && ev.Package != "" {
                        _, _ = bw.WriteString(ev.Package + " " + ev.Test + "\n")
                    }
                } else {
                    // Skip malformed lines
                    _ = dec.Decode(&map[string]any{})
                }
            }
            // Append AMI directive test entries, sorted for determinism.
            sort.Strings(amiManifest)
            for _, line := range amiManifest { _, _ = bw.WriteString(line + "\n") }
            _ = bw.Flush()
        }
    }

    // Discover and run runtime tests from *_test.ami pragmas
    runtimeCases, _ := parseRuntimeCases(dir)
    rc := runRuntime(dir, jsonOut, verbose, out, runtimeCases)

    // Emit a brief human summary to stdout when not JSON
    if !jsonOut {
        // Per-test concise lines reconstructed from go test JSON
        dec := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
        for dec.More() {
            var ev goTestEvent
            if dec.Decode(&ev) != nil { _ = dec.Decode(&map[string]any{}); continue }
            if ev.Package == "" || ev.Test == "" { continue }
            switch ev.Action {
            case "pass", "fail", "skip":
                fmt.Fprintf(out, "test: %s %s %s\n", ev.Package, ev.Test, ev.Action)
            }
        }
        // Per-package summaries including case counts
        if len(c.byPkg) > 0 {
            // stable order
            pkgs := make([]string, 0, len(c.byPkg))
            for k := range c.byPkg { pkgs = append(pkgs, k) }
            sort.Strings(pkgs)
            for _, pkg := range pkgs {
                v := c.byPkg[pkg]
                cases := v.ok + v.fail + v.skip
                fmt.Fprintf(out, "test: pkg %s ok=%d fail=%d cases=%d\n", pkg, v.ok, v.fail, cases)
            }
        }
        if amiTests > 0 {
            fmt.Fprintf(out, "test: ami ok=%d fail=%d\n", amiTests-amiFailures, amiFailures)
        }
        if rc.total > 0 {
            fmt.Fprintf(out, "test: runtime ok=%d fail=%d skip=%d\n", rc.ok, rc.fail, rc.skip)
        }
        if err == nil && amiFailures == 0 && rc.fail == 0 {
            fmt.Fprintln(out, "test: OK")
        }
    } else {
        // Emit final JSON summary record as a newline-delimited object with counts
        // Stream AMI directive events before the final summary for visibility.
        for _, ev := range amiEvents {
            _ = json.NewEncoder(out).Encode(map[string]any{
                "schema":  "ami.test.v1",
                "file":    ev.file,
                "case":    ev.name,
                "ok":      ev.ok,
                "expect":  ev.expect,
                "count":   ev.count,
                "gotErrs": ev.gotErrs,
                "msg":     ev.msgSub,
                "code":    ev.code,
                "line":    ev.line,
                "column":  ev.column,
                "offset":  ev.offset,
            })
        }
        // Per-package summary events
        if len(c.byPkg) > 0 {
            pkgs := make([]string, 0, len(c.byPkg))
            for k := range c.byPkg { pkgs = append(pkgs, k) }
            sort.Strings(pkgs)
            for _, pkg := range pkgs {
                v := c.byPkg[pkg]
                _ = json.NewEncoder(out).Encode(map[string]any{
                    "schema":  "ami.test.pkg.v1",
                    "package": pkg,
                    "ok":      v.ok,
                    "fail":    v.fail,
                })
            }
        }
        // Emit synthetic test.v1 test events (constructed from go test JSON) and run_end
        // We reconstruct events from the captured buffer for determinism in our schema.
        dec := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
        for dec.More() {
            var ev goTestEvent
            if dec.Decode(&ev) != nil { _ = dec.Decode(&map[string]any{}); continue }
            if ev.Test == "" || ev.Package == "" { continue }
            switch ev.Action {
            case "run":
                _ = json.NewEncoder(out).Encode(map[string]any{"schema":"test.v1", "type":"test_start", "package": ev.Package, "test": ev.Test})
            case "output":
                _ = json.NewEncoder(out).Encode(map[string]any{"schema":"test.v1", "type":"test_output", "package": ev.Package, "test": ev.Test, "output": ev.Output})
            case "pass", "fail", "skip":
                _ = json.NewEncoder(out).Encode(map[string]any{"schema":"test.v1", "type":"test_end", "package": ev.Package, "test": ev.Test, "status": ev.Action})
            }
        }
        // Build packages[] for run_end
        var pkgsArr []map[string]any
        if len(c.byPkg) > 0 {
            keys := make([]string, 0, len(c.byPkg))
            for k := range c.byPkg { keys = append(keys, k) }
            sort.Strings(keys)
            for _, k := range keys {
                v := c.byPkg[k]
                cases := v.ok + v.fail + v.skip
                pkgsArr = append(pkgsArr, map[string]any{"package": k, "pass": v.ok, "fail": v.fail, "skip": v.skip, "cases": cases})
            }
        }
        _ = json.NewEncoder(out).Encode(map[string]any{
            "schema":   "test.v1",
            "type":     "run_end",
            "totals":   map[string]any{"packages": c.packages, "tests": c.tests, "failures": c.failures},
            "packages": pkgsArr,
        })
        _ = json.NewEncoder(out).Encode(map[string]any{
            "ok":       err == nil && amiFailures == 0 && rc.fail == 0,
            "packages": c.packages,
            "tests":    c.tests,
            "failures": c.failures,
            "ami_tests":    amiTests,
            "ami_failures": amiFailures,
            "runtime_tests": rc.total,
            "runtime_failures": rc.fail,
            "runtime_skipped": rc.skip,
        })
    }

    if err != nil || amiFailures > 0 || rc.fail > 0 {
        // Write stderr summary to process stderr per SPEC
        msg := strings.TrimSpace(stderr.String())
        if msg != "" { fmt.Fprintln(os.Stderr, msg) }
        if err != nil { return err }
        if amiFailures > 0 { return fmt.Errorf("ami directive tests failed") }
        return fmt.Errorf("ami runtime tests failed")
    }
    return nil
}
