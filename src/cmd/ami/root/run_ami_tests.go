package root

import (
    "encoding/json"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "time"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/sem"
    kv "github.com/sam-caldwell/ami/src/ami/runtime/kvstore"
    "github.com/sam-caldwell/ami/src/ami/runtime/tester"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/internal/logger"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

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
        if len(roots) == 0 {
            roots = map[string]string{"main": "./src"}
        }
    }
    // Walk each root to find *_test.ami
    files := []string{}
    for _, root := range roots {
        filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
            if err != nil {
                return nil
            }
            if d.IsDir() {
                return nil
            }
            if strings.HasSuffix(d.Name(), "_test.ami") {
                files = append(files, path)
            }
            return nil
        })
    }
    sort.Strings(files)
    if len(files) == 0 {
        return
    }
    for _, file := range files {
        b, _ := os.ReadFile(file)
        src := string(b)
        p := parser.New(src)
        f := p.ParseFile()
        // derive cases from directives
        pkg := uPackageFromPath(file, roots)
        cases := deriveAmiCases(file, pkg, f)
        // fall back to default case when none were declared
        if len(cases) == 0 {
            cases = []amiCase{{name: filepath.Base(file), file: file, pkg: pkg, expects: []amiExpect{{kind: "no_errors"}}}}
        }
        // Run compiler front-end to collect diagnostics
        diags := sem.Check(f)
        if flagJSON {
            // emit per-case start, then end with status and optional details
            for _, c := range cases {
                s := sch.TestStart{Schema: "test.v1", Type: "test_start", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Package: pkg, Name: c.name}
                if s.Validate() == nil {
                    _ = json.NewEncoder(os.Stdout).Encode(s)
                }
                ok, details := evalAmiExpectations(c, diags)
                status := "pass"
                if !ok {
                    status = "fail"
                }
                e := sch.TestEnd{Schema: "test.v1", Type: "test_end", Timestamp: sch.FormatTimestamp(time.Now().UTC()), Package: pkg, Name: c.name, Status: status, DurationMs: 0}
                if e.Validate() == nil {
                    _ = json.NewEncoder(os.Stdout).Encode(e)
                }
                if !ok {
                    _ = details // in this phase, just ignore details in JSON stream
                }
            }
        } else {
            for _, c := range cases {
                ok, _ := evalAmiExpectations(c, diags)
                if ok {
                    *pass++
                    get(pkg).pass++
                    logger.Info(fmt.Sprintf("test PASS %s %s", pkg, c.name), nil)
                } else {
                    *fail++
                    get(pkg).fail++
                    logger.Error(fmt.Sprintf("test FAIL %s %s", pkg, c.name), nil)
                }
            }
        }

        // runtime tests (Phase 2 deterministic harness)
        // Gather runtime cases
        rcases := deriveAmiRuntimeCases(file, pkg, f)
        if len(rcases) > 0 {
            // execute via tester harness
            r := tester.New()
            if testKVAutoEmit {
                r.EnableAutoEmitKV(true)
            }
            // Bridge case types
            var cases2 []tester.Case
            for _, rc := range rcases {
                // fixtures (mode validation only in phase 2)
                var fx []tester.Fixture
                for _, it := range rc.fixtures {
                    fx = append(fx, tester.Fixture{Path: it.path, Mode: it.mode})
                }
                cases2 = append(cases2, tester.Case{Name: rc.name, Pipeline: rc.pipeline, InputJSON: rc.inputJSON, ExpectJSON: rc.expectJSON, ExpectError: rc.expectError, TimeoutMs: rc.timeoutMs, Fixtures: fx})
            }
            results, _ := r.Execute("", cases2)
            // Map results to events
            // start times used to compute duration when needed
            startTimes := map[string]time.Time{}
            resultsByKey := map[string]tester.Result{}
            key := func(pkg, name string) string { return pkg + ":" + name }
            for _, rc := range rcases {
                nm := rc.name
                if nm == "" {
                    nm = filepath.Base(file)
                }
                startTimes[key(pkg, nm)] = time.Now()
            }
            for _, r0 := range results {
                resultsByKey[key(pkg, r0.Name)] = r0
            }
            // Emit end events and update totals
            for _, rc := range rcases {
                name := rc.name
                if name == "" {
                    name = filepath.Base(file)
                }
                st := startTimes[key(pkg, name)]
                res, ok := resultsByKey[key(pkg, name)]
                if !ok {
                    res = tester.Result{Name: name, Status: "skip", Error: "runtime unavailable"}
                }
                status := res.Status
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
                    if e.Validate() == nil {
                        _ = json.NewEncoder(os.Stdout).Encode(e)
                    }
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
            // optional KV dump in human mode
            if testKVDump {
                infos := kv.Default().Snapshot()
                for _, inf := range infos {
                    if !flagJSON {
                        logger.Info("kvstore.dump "+inf.Pipeline+"/"+inf.Node, map[string]interface{}{"summary": inf.Stats, "dump": inf.Dump})
                    }
                }
            }
        }
    }
}

