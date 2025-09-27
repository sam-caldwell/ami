package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sort"
    "time"

    "github.com/sam-caldwell/ami/src/ami/exit"
    "github.com/sam-caldwell/ami/src/ami/compiler/driver"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    man "github.com/sam-caldwell/ami/src/ami/manifest"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// runBuild validates the workspace and prepares build configuration.
// For this phase, it enforces toolchain.* constraints and emits diagnostics.
func runBuild(out io.Writer, dir string, jsonOut bool, verbose bool) error {
    if lg := getRootLogger(); lg != nil {
        lg.Info("build.start", map[string]any{"dir": dir, "json": jsonOut})
    }

    wsPath := filepath.Join(dir, "ami.workspace")
    var ws workspace.Workspace
    st, err := os.Stat(wsPath)
    if errors.Is(err, os.ErrNotExist) || (err == nil && st.IsDir()) {
        // Missing or not a file: emit schema violation
        if jsonOut {
            rec := diag.Record{
                Timestamp: time.Now().UTC(),
                Level:     diag.Error,
                Code:      "E_WS_SCHEMA",
                Message:   "workspace validation failed: ami.workspace not found",
                File:      "ami.workspace",
            }
            _ = json.NewEncoder(out).Encode(rec)
        }
        return exit.New(exit.User, "workspace validation failed: ami.workspace not found")
    } else if err != nil {
        if jsonOut {
            rec := diag.Record{
                Timestamp: time.Now().UTC(),
                Level:     diag.Error,
                Code:      "E_WS_SCHEMA",
                Message:   fmt.Sprintf("workspace validation failed: stat error: %v", err),
                File:      "ami.workspace",
            }
            _ = json.NewEncoder(out).Encode(rec)
        }
        return exit.New(exit.IO, "stat workspace: %v", err)
    }

    if err := ws.Load(wsPath); err != nil {
        if jsonOut {
            rec := diag.Record{
                Timestamp: time.Now().UTC(),
                Level:     diag.Error,
                Code:      "E_WS_SCHEMA",
                Message:   fmt.Sprintf("workspace validation failed: load error: %v", err),
                File:      "ami.workspace",
            }
            _ = json.NewEncoder(out).Encode(rec)
        }
        return exit.New(exit.IO, "failed to load workspace: %v", err)
    }

    // Enforce schema-level constraints
    if errs := ws.Validate(); len(errs) > 0 {
        // Join errors into a single message string for summary
        msg := "workspace validation failed: " + errs[0]
        if len(errs) > 1 { msg += fmt.Sprintf(" (+%d more)", len(errs)-1) }
        if jsonOut {
            rec := diag.Record{
                Timestamp: time.Now().UTC(),
                Level:     diag.Error,
                Code:      "E_WS_SCHEMA",
                Message:   msg,
                File:      "ami.workspace",
                Data:      map[string]any{"errors": errs},
            }
            _ = json.NewEncoder(out).Encode(rec)
        }
        return exit.New(exit.User, "%s", msg)
    }

    // Configuration from workspace
    // - target directory (workspace-relative; validated by ws.Validate)
    target := ws.Toolchain.Compiler.Target
    if target == "" { target = "./build" }
    absTarget := filepath.Clean(filepath.Join(dir, target))

    // - env matrix (default if empty per phase scope)
    envs := ws.Toolchain.Compiler.Env
    if len(envs) == 0 {
        envs = []string{"darwin/arm64"}
        if lg := getRootLogger(); lg != nil {
            lg.Info("build.env.default", map[string]any{"env": envs[0]})
        }
    }

    // For this phase, stop after validation.
    // Enforce dependency availability per workspace requirements (scaffold via audit).
    rep, aerr := workspace.AuditDependencies(dir)
    if aerr != nil {
        if jsonOut {
            rec := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_INTEGRITY", Message: fmt.Sprintf("dependency audit failed: %v", aerr), File: "ami.workspace"}
            _ = json.NewEncoder(out).Encode(rec)
        }
        return exit.New(exit.IO, "dependency audit failed: %v", aerr)
    }

    // If ami.manifest exists alongside ami.sum, verify consistency with sum.
    // Compare presence of name@version pairs irrespective of sha.
    maniPath := filepath.Join(dir, "ami.manifest")
    if st, err := os.Stat(maniPath); err == nil && !st.IsDir() {
        sumPath := filepath.Join(dir, "ami.sum")
        if st2, err2 := os.Stat(sumPath); err2 == nil && !st2.IsDir() {
            // load sum
            var sum workspace.Manifest
            if err := sum.Load(sumPath); err == nil {
                // load ami.manifest
                var mf man.Manifest
                if err := mf.Load(maniPath); err == nil {
                    // extract name@version set from mf.Data["packages"] when possible
                    inMani := make(map[string]struct{})
                    if pk, ok := mf.Data["packages"]; ok {
                        switch t := pk.(type) {
                        case map[string]any:
                            for name, v := range t {
                                if mm, ok := v.(map[string]any); ok {
                                    for ver := range mm { inMani[name+"@"+ver] = struct{}{} }
                                }
                            }
                        case []any:
                            for _, el := range t {
                                if mm, ok := el.(map[string]any); ok {
                                    name, _ := mm["name"].(string)
                                    ver, _ := mm["version"].(string)
                                    if name != "" && ver != "" { inMani[name+"@"+ver] = struct{}{} }
                                }
                            }
                        }
                    }
                    // extract from sum
                    inSum := make(map[string]struct{})
                    for name, vv := range sum.Packages { for ver := range vv { inSum[name+"@"+ver] = struct{}{} } }
                    // compute deltas
                    var missingInMani, extraInMani []string
                    for k := range inSum { if _, ok := inMani[k]; !ok { missingInMani = append(missingInMani, k) } }
                    for k := range inMani { if _, ok := inSum[k]; !ok { extraInMani = append(extraInMani, k) } }
                    if len(missingInMani) > 0 || len(extraInMani) > 0 {
                        if jsonOut {
                            _ = json.NewEncoder(out).Encode(diag.Record{
                                Timestamp: time.Now().UTC(),
                                Level:     diag.Error,
                                Code:      "E_INTEGRITY_MANIFEST",
                                Message:   "ami.manifest mismatch vs ami.sum",
                                File:      "ami.manifest",
                                Data:      map[string]any{"missing": missingInMani, "extra": extraInMani},
                            })
                        }
                        return exit.New(exit.Integrity, "manifest mismatch with ami.sum")
                    }
                }
            }
        }
    }
    // Consider issues when ami.sum missing or any lists are non-empty as violations for this phase.
    if len(rep.Requirements) > 0 && (!rep.SumFound || len(rep.MissingInSum) > 0 || len(rep.Unsatisfied) > 0 || len(rep.MissingInCache) > 0 || len(rep.Mismatched) > 0 || len(rep.ParseErrors) > 0) {
        if jsonOut {
            enc := json.NewEncoder(out)
            now := time.Now().UTC()
            // Emit per-item diagnostics for cache integrity mismatches to aid tooling.
            for _, k := range rep.MissingInCache {
                _ = enc.Encode(diag.Record{
                    Timestamp: now,
                    Level:     diag.Error,
                    Code:      "E_INTEGRITY",
                    Message:   "dependency missing from cache: " + k,
                    File:      "ami.sum",
                    Data:      map[string]any{"kind": "missingInCache", "key": k},
                })
            }
            for _, k := range rep.Mismatched {
                _ = enc.Encode(diag.Record{
                    Timestamp: now,
                    Level:     diag.Error,
                    Code:      "E_INTEGRITY",
                    Message:   "dependency hash mismatch: " + k,
                    File:      "ami.sum",
                    Data:      map[string]any{"kind": "mismatched", "key": k},
                })
            }
            // Emit a summary record last with full context for consumers.
            _ = enc.Encode(diag.Record{
                Timestamp: now,
                Level:     diag.Error,
                Code:      "E_INTEGRITY",
                Message:   "dependency integrity check failed; run 'ami mod update'",
                File:      "ami.workspace",
                Data: map[string]any{
                    "sumFound":       rep.SumFound,
                    "missingInSum":   rep.MissingInSum,
                    "unsatisfied":    rep.Unsatisfied,
                    "missingInCache": rep.MissingInCache,
                    "mismatched":     rep.Mismatched,
                    "parseErrors":    rep.ParseErrors,
                },
            })
        }
        return exit.New(exit.Integrity, "dependency integrity check failed; run 'ami mod update'")
    }

    // When verbose, emit front-end debug artifacts (AST/IR/etc.) and build plan including obj index.
    if verbose {
        _ = os.MkdirAll(filepath.Join(dir, "build", "debug"), 0o755)
        // Front-end debug artifacts: parse main package .ami files and compile with Debug=true
        if p := ws.FindPackage("main"); p != nil && p.Root != "" {
            root := filepath.Clean(filepath.Join(dir, p.Root))
            // check for missing package root
            if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
                if jsonOut {
                    rec := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_FS_MISSING", Message: fmt.Sprintf("missing package root: %s", root), File: "ami.workspace"}
                    _ = json.NewEncoder(out).Encode(rec)
                }
                return exit.New(exit.IO, "missing package root: %s", root)
            }
            // Collect .ami files under root
            var files []string
            _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
                if err != nil || d.IsDir() { return nil }
                if filepath.Ext(path) == ".ami" { files = append(files, path) }
                return nil
            })
            if len(files) > 0 {
                var fs source.FileSet
                for _, f := range files {
                    b, err := os.ReadFile(f); if err != nil { continue }
                    fs.AddFile(f, string(b))
                }
                // Run compile with CWD set to workspace dir so relative debug paths land under dir/build/debug
                oldwd, _ := os.Getwd()
                _ = os.Chdir(dir)
                pkgs := []driver.Package{{Name: p.Name, Files: &fs}}
                // hook logger for full timestamped compiler activity under build/debug/activity.log
                var logcb func(string, map[string]any)
                if lg := getRootLogger(); lg != nil {
                    logcb = func(event string, fields map[string]any) { lg.Info("compiler."+event, fields) }
                }
                _, _ = driver.Compile(ws, pkgs, driver.Options{Debug: true, Log: logcb})
                _ = os.Chdir(oldwd)
            }
        }
        // Build plan after emitting artifacts; include object index paths when present
        planPath := filepath.Join(dir, "build", "debug", "build.plan.json")
        type planPkg struct{ Key, Name, Version, Root string }
        plan := struct {
            Schema    string    `json:"schema"`
            TargetDir string    `json:"targetDir"`
            Targets   []string  `json:"targets"`
            Packages  []planPkg `json:"packages"`
            ObjIndex  []string  `json:"objIndex,omitempty"`
        }{Schema: "build.plan/v1", TargetDir: absTarget, Targets: envs}
        for _, e := range ws.Packages {
            plan.Packages = append(plan.Packages, planPkg{Key: e.Key, Name: e.Package.Name, Version: e.Package.Version, Root: e.Package.Root})
            // if object index exists for this package, include path
            idx := filepath.Join(dir, "build", "obj", e.Package.Name, "index.json")
            if st, err := os.Stat(idx); err == nil && !st.IsDir() {
                rel, _ := filepath.Rel(dir, idx)
                plan.ObjIndex = append(plan.ObjIndex, rel)
            }
        }
        if f, err := os.OpenFile(planPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644); err == nil {
            _ = json.NewEncoder(f).Encode(plan)
            _ = f.Close()
        }
    }

    // In JSON mode, run a non-debug compile to surface parser/semantic diagnostics as a stream.
    if jsonOut {
        if p := ws.FindPackage("main"); p != nil && p.Root != "" {
            root := filepath.Clean(filepath.Join(dir, p.Root))
            var files []string
            _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
                if err != nil || d.IsDir() { return nil }
                if filepath.Ext(path) == ".ami" { files = append(files, path) }
                return nil
            })
            if len(files) > 0 {
                var fs source.FileSet
                for _, f := range files { b, err := os.ReadFile(f); if err == nil { fs.AddFile(f, string(b)) } }
                pkgs := []driver.Package{{Name: p.Name, Files: &fs}}
                var logcb func(string, map[string]any)
                if lg := getRootLogger(); lg != nil {
                    logcb = func(event string, fields map[string]any) { lg.Info("compiler."+event, fields) }
                }
                _, diags := driver.Compile(ws, pkgs, driver.Options{Debug: false, Log: logcb})
                if len(diags) > 0 {
                    enc := json.NewEncoder(out)
                    for i := range diags { _ = enc.Encode(diags[i]) }
                    return exit.New(exit.User, "compiler reported diagnostics")
                }
            }
        }
    }

    // Always perform a non-debug compile pass to emit object stubs + object index under build/obj
    if p := ws.FindPackage("main"); p != nil && p.Root != "" {
        root := filepath.Clean(filepath.Join(dir, p.Root))
        var files []string
        _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
            if err != nil || d.IsDir() { return nil }
            if filepath.Ext(path) == ".ami" { files = append(files, path) }
            return nil
        })
        if len(files) > 0 {
            var fs source.FileSet
            for _, f := range files { b, err := os.ReadFile(f); if err == nil { fs.AddFile(f, string(b)) } }
            // Run compile with CWD set to workspace dir so outputs land under dir/build
            oldwd, _ := os.Getwd()
            _ = os.Chdir(dir)
            pkgs := []driver.Package{{Name: p.Name, Files: &fs}}
            _, diags := driver.Compile(ws, pkgs, driver.Options{Debug: false})
            _ = os.Chdir(oldwd)
            if jsonOut && len(diags) > 0 {
                enc := json.NewEncoder(out)
                for i := range diags { _ = enc.Encode(diags[i]) }
                return exit.New(exit.User, "compiler reported diagnostics")
            }
        }
    }

    // Rewrite ami.manifest into build/ami.manifest with toolchain metadata and objIndex entries.
    {
        buildDir := filepath.Join(dir, "build")
        _ = os.MkdirAll(buildDir, 0o755)
        objIdx := []string{}
        for _, e := range ws.Packages {
            idx := filepath.Join(dir, "build", "obj", e.Package.Name, "index.json")
            if st, err := os.Stat(idx); err == nil && !st.IsDir() {
                rel, _ := filepath.Rel(dir, idx)
                objIdx = append(objIdx, rel)
            }
        }
        // Load ami.sum if present to embed package evidence.
        sumPath := filepath.Join(dir, "ami.sum")
        pkgs := map[string]map[string]string{}
        var sum workspace.Manifest
        if st, err := os.Stat(sumPath); err == nil && !st.IsDir() {
            if err := sum.Load(sumPath); err == nil { pkgs = sum.Packages }
        }
        sort.Strings(objIdx)
        outObj := map[string]any{
            "schema":    "ami.manifest/v1",
            "packages":  pkgs,
            "toolchain": map[string]any{"targetDir": absTarget, "targets": envs},
            "objIndex":  objIdx,
        }
        // integrity evidence from ami.sum vs cache
        if len(sum.Packages) > 0 {
            if v, m, mm, err := sum.Validate(); err == nil {
                outObj["integrity"] = map[string]any{"verified": v, "missing": m, "mismatched": mm}
            }
        }
        // discover binaries under build/ (exclude debug/ and obj/); treat executable regular files as binaries
        var bins []string
        _ = filepath.WalkDir(buildDir, func(path string, d os.DirEntry, err error) error {
            if err != nil { return nil }
            if d.IsDir() {
                // skip debug and obj subtrees
                base := filepath.Base(path)
                if base == "debug" || base == "obj" { return filepath.SkipDir }
                return nil
            }
            // regular file: check any execute bit
            if info, e := d.Info(); e == nil {
                mode := info.Mode()
                if mode.IsRegular() && (mode&0o111 != 0) {
                    if rel, rerr := filepath.Rel(dir, path); rerr == nil { bins = append(bins, rel) }
                }
            }
            return nil
        })
        if len(bins) > 0 {
            sort.Strings(bins)
            outObj["binaries"] = bins
        }
        if verbose {
            // collect debug artifact references for cross-linking
            var debugRefs []string
            // AST
            for _, e := range ws.Packages {
                glob := filepath.Join(dir, "build", "debug", "ast", e.Package.Name, "*.ast.json")
                if matches, _ := filepath.Glob(glob); len(matches) > 0 {
                    for _, m := range matches { if rel, err := filepath.Rel(dir, m); err == nil { debugRefs = append(debugRefs, rel) } }
                }
                // IR
                glob = filepath.Join(dir, "build", "debug", "ir", e.Package.Name, "*.ir.json")
                if matches, _ := filepath.Glob(glob); len(matches) > 0 {
                    for _, m := range matches { if rel, err := filepath.Rel(dir, m); err == nil { debugRefs = append(debugRefs, rel) } }
                }
                // ASM listings
                glob = filepath.Join(dir, "build", "debug", "asm", e.Package.Name, "*.s")
                if matches, _ := filepath.Glob(glob); len(matches) > 0 {
                    for _, m := range matches { if rel, err := filepath.Rel(dir, m); err == nil { debugRefs = append(debugRefs, rel) } }
                }
                // Edges index
                idx := filepath.Join(dir, "build", "debug", "asm", e.Package.Name, "edges.json")
                if st, err := os.Stat(idx); err == nil && !st.IsDir() { if rel, err := filepath.Rel(dir, idx); err == nil { debugRefs = append(debugRefs, rel) } }
            }
            if len(debugRefs) > 0 { sort.Strings(debugRefs); outObj["debug"] = debugRefs }
        }
        f, err := os.OpenFile(filepath.Join(buildDir, "ami.manifest"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
        if err == nil {
            _ = json.NewEncoder(f).Encode(outObj)
            _ = f.Close()
        }
    }

    if jsonOut {
        // Collect object index paths when available (verbose compile may have produced them)
        var objIdx []string
        for _, e := range ws.Packages {
            idx := filepath.Join(dir, "build", "obj", e.Package.Name, "index.json")
            if st, err := os.Stat(idx); err == nil && !st.IsDir() {
                rel, _ := filepath.Rel(dir, idx)
                objIdx = append(objIdx, rel)
            }
        }
        sort.Strings(objIdx)
        // Emit a simple success summary for consistency with machine parsing.
        rec := diag.Record{
            Timestamp: time.Now().UTC(),
            Level:     diag.Info,
            Code:      "BUILD_OK",
            Message:   "workspace valid; build planning deferred",
            File:      "ami.workspace",
            Data:      map[string]any{"targets": envs, "targetDir": absTarget, "objIndex": objIdx, "buildManifest": filepath.Join("build", "ami.manifest")},
        }
        return json.NewEncoder(out).Encode(rec)
    }
    // Human output kept minimal per SPEC I/O rules.
    fmt.Fprintf(out, "workspace valid: target=%s envs=%d\n", absTarget, len(envs))
    return nil
}
