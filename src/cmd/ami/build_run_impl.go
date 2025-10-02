package main

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/codegen"
    "github.com/sam-caldwell/ami/src/ami/compiler/driver"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/exit"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/ami/runtime/kvstore"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// runBuildImpl validates the workspace and prepares build configuration.
// For this phase, it enforces toolchain.* constraints and emits diagnostics.
func runBuildImpl(out io.Writer, dir string, jsonOut bool, verbose bool) error {
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
        if len(errs) > 1 {
            msg += fmt.Sprintf(" (+%d more)", len(errs)-1)
        }
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

    // Enforce local cross-package composition contracts (local imports):
    // - Imported local path must exist
    // - Imported local path should be declared as a workspace package root
    {
        // Build a set of declared roots for quick lookup
        declaredRoots := map[string]bool{}
        for _, e := range ws.Packages {
            root := filepath.Clean(filepath.Join(dir, strings.TrimSpace(e.Package.Root)))
            declaredRoots[root] = true
        }
        for _, e := range ws.Packages {
            p := e.Package
            workspace.NormalizeImports(&p)
            for _, ent := range p.Import {
                path, _ := workspace.ParseImportEntry(ent)
                if path == "" || !strings.HasPrefix(path, "./") {
                    continue
                }
                abs := filepath.Clean(filepath.Join(dir, path))
                if st, err := os.Stat(abs); errors.Is(err, os.ErrNotExist) || (err == nil && !st.IsDir()) {
                    if jsonOut {
                        _ = json.NewEncoder(out).Encode(diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_IMPORT_LOCAL_MISSING", Message: "local import path not found: " + path, File: "ami.sum", Data: map[string]any{"package": p.Name, "import": path}})
                    }
                    return exit.New(exit.User, "local import path not found: %s", path)
                }
                if !declaredRoots[abs] {
                    if jsonOut {
                        _ = json.NewEncoder(out).Encode(diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_IMPORT_LOCAL_UNDECLARED", Message: "local import not declared as package: " + path, File: "ami.workspace", Data: map[string]any{"package": p.Name, "import": path}})
                    }
                    return exit.New(exit.User, "local import not declared as package: %s", path)
                }
            }
        }
    }

    // Verify package root paths exist; emit diag when missing
    for _, e := range ws.Packages {
        if e.Package.Root == "" || e.Package.Root == "./src" { continue }
        root := filepath.Clean(filepath.Join(dir, e.Package.Root))
        if st, err := os.Stat(root); errors.Is(err, os.ErrNotExist) || (err == nil && !st.IsDir()) {
            if jsonOut {
                _ = json.NewEncoder(out).Encode(diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_FS_MISSING", Message: "package root missing: " + e.Package.Root, File: "ami.workspace"})
            }
            return exit.New(exit.IO, "package root missing: %s", e.Package.Root)
        }
    }

    // Configuration from workspace
    // - target directory (workspace-relative; validated by ws.Validate)
    target := ws.Toolchain.Compiler.Target
    if target == "" {
        target = "./build"
    }
    absTarget := filepath.Clean(filepath.Join(dir, target))

    // - env matrix (default if empty per phase scope). When present, strictly honor envs.
    envs := ws.Toolchain.Compiler.Env
    if len(envs) == 0 {
        envs = []string{"darwin/arm64"}
        if lg := getRootLogger(); lg != nil {
            lg.Info("build.env.default", map[string]any{"env": envs[0]})
        }
    }
    // Ensure downstream compile sees the resolved env matrix
    ws.Toolchain.Compiler.Env = envs

    // - backend selection: CLI flag overrides workspace value
    backendName := ws.Toolchain.Compiler.Backend
    if buildBackend != "" {
        backendName = buildBackend
    }
    if backendName == "" { backendName = "llvm" }
    // Apply backend selection globally
    _ = codegen.SelectDefaultBackend(backendName)


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

    // If ami.manifest exists alongside ami.sum, perform a simple presence check
    maniPath := filepath.Join(dir, "ami.manifest")
    if st, err := os.Stat(maniPath); err == nil && !st.IsDir() {
        var sum workspace.Manifest
        var mani workspace.Manifest
        sumPath := filepath.Join(dir, "ami.sum")
        if st, err := os.Stat(sumPath); err == nil && !st.IsDir() { _ = sum.Load(sumPath) }
        _ = mani.Load(maniPath)
        type pair struct{ n, v string }
        have := map[pair]bool{}
        want := map[pair]bool{}
        for name, vers := range sum.Packages { for ver := range vers { have[pair{name, ver}] = true } }
        for name, vers := range mani.Packages { for ver := range vers { want[pair{name, ver}] = true } }
        mismatch := false
        if len(have) != len(want) { mismatch = true }
        if !mismatch {
            for k := range have { if !want[k] { mismatch = true; break } }
            if !mismatch { for k := range want { if !have[k] { mismatch = true; break } } }
        }
        if mismatch {
            if jsonOut {
                _ = json.NewEncoder(out).Encode(diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_INTEGRITY_MANIFEST", Message: "ami.manifest disagrees with ami.sum", File: "ami.manifest", Data: map[string]any{"sum": sum.Packages, "mani": mani.Packages}})
            }
            return exit.New(exit.Integrity, "manifest mismatch")
        }
    }

    // Optional signature verification: if ami.sum.sig or ami.manifest.sig exists, verify sha256 digest
    verifySig := func(file, sigFile string) error {
        b, err := os.ReadFile(file); if err != nil { return err }
        sum := sha256.Sum256(b)
        wantHex := hex.EncodeToString(sum[:])
        sigb, err := os.ReadFile(sigFile); if err != nil { return err }
        got := strings.TrimSpace(string(sigb))
        if !strings.EqualFold(got, wantHex) {
            if jsonOut {
                _ = json.NewEncoder(out).Encode(diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_INTEGRITY_SIGNATURE", Message: fmt.Sprintf("signature mismatch for %s", filepath.Base(file)), File: filepath.Base(sigFile), Data: map[string]any{"expected": wantHex, "got": got}})
            }
            return exit.New(exit.Integrity, "signature mismatch: %s", filepath.Base(file))
        }
        return nil
    }
    if _, err := os.Stat(filepath.Join(dir, "ami.sum.sig")); err == nil {
        if err := verifySig(filepath.Join(dir, "ami.sum"), filepath.Join(dir, "ami.sum.sig")); err != nil { return err }
    }
    if _, err := os.Stat(filepath.Join(dir, "ami.manifest.sig")); err == nil {
        if err := verifySig(filepath.Join(dir, "ami.manifest"), filepath.Join(dir, "ami.manifest.sig")); err != nil { return err }
    }

    // Consider issues when ami.sum missing or any lists are non-empty as violations for this phase.
    if len(rep.Requirements) > 0 && (!rep.SumFound || len(rep.MissingInSum) > 0 || len(rep.Unsatisfied) > 0 || len(rep.MissingInCache) > 0 || len(rep.Mismatched) > 0 || len(rep.ParseErrors) > 0) {
        if jsonOut {
            enc := json.NewEncoder(out)
            now := time.Now().UTC()
            for _, k := range rep.MissingInCache { _ = enc.Encode(diag.Record{Timestamp: now, Level: diag.Error, Code: "E_INTEGRITY", Message: "dependency missing from cache: " + k, File: "ami.sum", Data: map[string]any{"kind": "missingInCache", "key": k}}) }
            for _, k := range rep.Mismatched { _ = enc.Encode(diag.Record{Timestamp: now, Level: diag.Error, Code: "E_INTEGRITY", Message: "dependency hash mismatch: " + k, File: "ami.sum", Data: map[string]any{"kind": "mismatched", "key": k}}) }
            _ = enc.Encode(diag.Record{Timestamp: now, Level: diag.Error, Code: "E_INTEGRITY", Message: "dependency integrity check failed; run 'ami mod update'", File: "ami.workspace", Data: map[string]any{"sumFound": rep.SumFound, "missingInSum": rep.MissingInSum, "unsatisfied": rep.Unsatisfied, "missingInCache": rep.MissingInCache, "mismatched": rep.Mismatched, "parseErrors": rep.ParseErrors, "requirements": rep.Requirements}})
        }
        return exit.New(exit.Integrity, "dependency integrity check failed; run 'ami mod update'")
    }

    // Build phase per package (codegen/driver) and write artifacts in build/
    objRoot := filepath.Join(dir, "build", "obj")
    _ = os.MkdirAll(objRoot, 0o755)
    // Build debug: ensure debug dir exists when verbose (placeholder for future debug streams)
    if verbose {
        _ = os.MkdirAll(filepath.Join(dir, "build", "debug"), 0o755)
    }

    // Per-package compile loop
    var pkgs []driver.Package
    for _, e := range ws.Packages {
        p := e.Package
        // Skip packages lacking any .ami files
        pdir := filepath.Clean(filepath.Join(dir, p.Root))
        var files []string
        _ = filepath.WalkDir(pdir, func(path string, d os.DirEntry, err error) error {
            if err != nil || d.IsDir() { return nil }
            if filepath.Ext(path) == ".ami" { files = append(files, path) }
            return nil
        })
        if len(files) == 0 { continue }
        // Prepare output folder
        pkgObj := filepath.Join(objRoot, p.Name)
        _ = os.MkdirAll(pkgObj, 0o755)
        _ = func() error {
            idxPath := filepath.Join(pkgObj, "index.json")
            f, err := os.OpenFile(idxPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
            if err == nil {
                _ = json.NewEncoder(f).Encode(map[string]any{
                    "schema":  "ami.obj.index/v1",
                    "package": p.Name,
                    "files":   files,
                })
                _ = f.Close()
            }
            return nil
        }()
        // Per-file object emission via driver is no longer used here; driver.Compile aggregates per-package below.
        // Note: legacy error pipeline collection disabled; compiler errors are surfaced by later stages.
        // Aggregate package into driver compile phase
        _ = func() error {
            var files []string
            _ = filepath.WalkDir(pdir, func(path string, d os.DirEntry, err error) error {
                if err != nil || d.IsDir() { return nil }
                if filepath.Ext(path) == ".ami" { files = append(files, path) }
                return nil
            })
            if len(files) == 0 { return nil }
            var fs source.FileSet
            for _, f := range files { if b, err := os.ReadFile(f); err == nil { fs.AddFile(f, string(b)) } }
            pkgs = append(pkgs, driver.Package{Name: p.Name, Files: &fs})
            return nil
        }()
    }
    if len(pkgs) > 0 {
        oldwd, _ := os.Getwd()
        _ = os.Chdir(dir)
        var logFn func(string, map[string]any)
        if verbose {
            logFn = func(event string, fields map[string]any) {
                if lg := getRootLogger(); lg != nil { lg.Info("compiler."+event, fields) }
            }
        }
        _, diags := driver.Compile(ws, pkgs, driver.Options{Debug: verbose, EmitLLVMOnly: buildEmitLLVMOnly, NoLink: buildNoLink, Log: logFn})
        _ = os.Chdir(oldwd)
        if jsonOut && len(diags) > 0 {
            enc := json.NewEncoder(out)
            for _, d := range diags { _ = enc.Encode(d) }
            return exit.New(exit.User, "%s", "compile diagnostics")
        }
    }

    // Verbose: emit kvstore metrics and dump under build/debug/kv/
    if verbose {
        kvDir := filepath.Join(dir, "build", "debug", "kv")
        _ = os.MkdirAll(kvDir, 0o755)
        st := kvstore.Default()
        mts := st.Metrics()
        _ = writeJSONFile(filepath.Join(kvDir, "metrics.json"), map[string]any{"schema": "kv.metrics.v1", "hits": mts.Hits, "misses": mts.Misses, "expirations": mts.Expirations, "evictions": mts.Evictions, "currentSize": mts.CurrentSize})
        keys := st.Keys()
        _ = writeJSONFile(filepath.Join(kvDir, "dump.json"), map[string]any{"schema": "kv.dump.v1", "keys": keys, "size": len(keys)})
    }

    if !buildNoLink { buildLink(out, dir, ws, envs, jsonOut) }

    // Rewrite ami.manifest into build/ami.manifest with toolchain metadata and objIndex entries.
    {
        buildDir := filepath.Join(dir, "build")
        _ = os.MkdirAll(buildDir, 0o755)
        objIdx := []string{}
        for _, e := range ws.Packages {
            idx := filepath.Join(dir, "build", "obj", e.Package.Name, "index.json")
            if st, err := os.Stat(idx); err == nil && !st.IsDir() { if rel, _ := filepath.Rel(dir, idx); rel != "" { objIdx = append(objIdx, rel) } }
        }
        // Load ami.sum if present to embed package evidence.
        sumPath := filepath.Join(dir, "ami.sum")
        pkgsMap := map[string]map[string]string{}
        var sum workspace.Manifest
        if st, err := os.Stat(sumPath); err == nil && !st.IsDir() { if err := sum.Load(sumPath); err == nil { pkgsMap = sum.Packages } }
        sort.Strings(objIdx)
        outObj := map[string]any{"schema": "ami.manifest/v1", "packages": pkgsMap, "toolchain": map[string]any{"targetDir": absTarget, "targets": envs}, "objIndex": objIdx}
        // Embed simple integrity evidence: versions present in sum that are neither missing nor mismatched in cache
        if pkgsMap != nil {
            bad := map[string]bool{}
            for _, k := range rep.MissingInCache { bad[k] = true }
            for _, k := range rep.Mismatched { bad[k] = true }
            var verified []string
            for name, vers := range pkgsMap {
                for ver := range vers {
                    key := name + "@" + ver
                    if !bad[key] { verified = append(verified, key) }
                }
            }
            if len(verified) > 0 {
                outObj["integrity"] = map[string]any{"verified": verified}
            }
        }
        var objects []string
        var artifacts []map[string]any
        for _, e := range ws.Packages {
            glob := filepath.Join(dir, "build", "obj", e.Package.Name, "*.o")
            if matches, _ := filepath.Glob(glob); len(matches) > 0 {
                for _, m := range matches { if rel, err := filepath.Rel(dir, m); err == nil { objects = append(objects, rel); artifacts = append(artifacts, map[string]any{"kind": "obj", "path": rel}) } }
            }
        }
        if len(objects) > 0 { sort.Strings(objects); outObj["objects"] = objects }
        var rtObjs []string
        for _, env := range envs {
            glob := filepath.Join(dir, "build", "runtime", env, "*.o")
            if matches, _ := filepath.Glob(glob); len(matches) > 0 {
                for _, m := range matches { if rel, err := filepath.Rel(dir, m); err == nil { rtObjs = append(rtObjs, rel); artifacts = append(artifacts, map[string]any{"kind": "runtime_obj", "path": rel, "env": env}) } }
            }
        }
        if len(rtObjs) > 0 { sort.Strings(rtObjs); outObj["runtimeObjects"] = rtObjs }
        if len(artifacts) > 0 { outObj["artifacts"] = artifacts }
        var bins []string
        _ = filepath.WalkDir(buildDir, func(path string, d os.DirEntry, err error) error {
            if err != nil { return nil }
            if d.IsDir() { base := filepath.Base(path); if base == "debug" || base == "obj" { return filepath.SkipDir }; return nil }
            if info, e := d.Info(); e == nil { mode := info.Mode(); if mode.IsRegular() && (mode&0o111 != 0) { if rel, rerr := filepath.Rel(dir, path); rerr == nil { bins = append(bins, rel) } } }
            return nil
        })
        if len(bins) > 0 { sort.Strings(bins); outObj["binaries"] = bins }
        binsByEnv := map[string][]string{}
        for _, env := range envs {
            envDir := filepath.Join(buildDir, env)
            var list []string
            _ = filepath.WalkDir(envDir, func(path string, d os.DirEntry, err error) error {
                if err != nil { return nil }
                if d.IsDir() { return nil }
                if info, e := d.Info(); e == nil { mode := info.Mode(); if mode.IsRegular() && (mode&0o111 != 0) { if rel, rerr := filepath.Rel(dir, path); rerr == nil { list = append(list, rel) } } }
                return nil
            })
            if len(list) > 0 { sort.Strings(list); binsByEnv[env] = list }
        }
        if len(binsByEnv) > 0 { outObj["binariesByEnv"] = binsByEnv }
        objectsByEnv := map[string][]string{}
        objIndexByEnv := map[string][]string{}
        for _, env := range envs {
            var objs []string
            for _, e := range ws.Packages {
                glob := filepath.Join(dir, "build", env, "obj", e.Package.Name, "*.o")
                if matches, _ := filepath.Glob(glob); len(matches) > 0 { for _, m := range matches { if rel, err := filepath.Rel(dir, m); err == nil { objs = append(objs, rel) } } }
                idx := filepath.Join(dir, "build", env, "obj", e.Package.Name, "index.json")
                if st, err := os.Stat(idx); err == nil && !st.IsDir() { if rel, rerr := filepath.Rel(dir, idx); rerr == nil { objIndexByEnv[env] = append(objIndexByEnv[env], rel) } }
            }
            if len(objs) > 0 { sort.Strings(objs); objectsByEnv[env] = objs }
        }
        if len(objectsByEnv) > 0 { outObj["objectsByEnv"] = objectsByEnv }
        if len(objIndexByEnv) > 0 { outObj["objIndexByEnv"] = objIndexByEnv }
        if verbose {
            var debugRefs []string
            for _, e := range ws.Packages {
                glob := filepath.Join(dir, "build", "debug", "ast", e.Package.Name, "*.ast.json")
                if matches, _ := filepath.Glob(glob); len(matches) > 0 { for _, m := range matches { if rel, err := filepath.Rel(dir, m); err == nil { debugRefs = append(debugRefs, rel) } } }
                glob = filepath.Join(dir, "build", "debug", "ir", e.Package.Name, "*.ir.json")
                if matches, _ := filepath.Glob(glob); len(matches) > 0 { for _, m := range matches { if rel, err := filepath.Rel(dir, m); err == nil { debugRefs = append(debugRefs, rel) } } }
                glob = filepath.Join(dir, "build", "debug", "asm", e.Package.Name, "*.s")
                if matches, _ := filepath.Glob(glob); len(matches) > 0 { for _, m := range matches { if rel, err := filepath.Rel(dir, m); err == nil { debugRefs = append(debugRefs, rel) } } }
                idx := filepath.Join(dir, "build", "debug", "asm", e.Package.Name, "edges.json")
                if st, err := os.Stat(idx); err == nil && !st.IsDir() { if rel, err := filepath.Rel(dir, idx); err == nil { debugRefs = append(debugRefs, rel) } }
            }
            if len(debugRefs) > 0 { sort.Strings(debugRefs); outObj["debug"] = debugRefs }
        }
        if f, err := os.OpenFile(filepath.Join(buildDir, "ami.manifest"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644); err == nil {
            _ = json.NewEncoder(f).Encode(outObj)
            _ = f.Close()
        }
    }

    // Write verbose build plan after artifacts are emitted
    if verbose {
        planPath := filepath.Join(dir, "build", "debug", "build.plan.json")
        type planPkg struct { Key, Name, Version, Root string; HasObjects bool `json:"hasObjects"` }
        plan := struct {
            Schema    string     `json:"schema"`
            TargetDir string     `json:"targetDir"`
            Targets   []string   `json:"targets"`
            Packages  []planPkg  `json:"packages"`
            ObjIndex  []string   `json:"objIndex,omitempty"`
            ObjectsByEnv map[string][]string `json:"objectsByEnv,omitempty"`
            ObjIndexByEnv map[string][]string `json:"objIndexByEnv,omitempty"`
            Objects []string `json:"objects,omitempty"`
        }{ Schema: "build.plan/v1", TargetDir: absTarget, Targets: envs }
        for _, e := range ws.Packages {
            root := filepath.Clean(filepath.Join(dir, e.Package.Root))
            key := e.Package.Name
            if e.Package.Version != "" { key = key + "@" + e.Package.Version }
            hasObj := false
            if matches, _ := filepath.Glob(filepath.Join(dir, "build", "obj", e.Package.Name, "*.o")); len(matches) > 0 { hasObj = true }
            plan.Packages = append(plan.Packages, planPkg{Key: key, Name: e.Package.Name, Version: e.Package.Version, Root: root, HasObjects: hasObj})
        }
        sort.Slice(plan.Packages, func(i, j int) bool { return plan.Packages[i].Name < plan.Packages[j].Name })
        // Populate objIndex and objectsByEnv
        var idx []string
        objsByEnv := map[string][]string{}
        for _, e := range ws.Packages {
            p := filepath.Join(dir, "build", "obj", e.Package.Name, "index.json")
            if st, err := os.Stat(p); err == nil && !st.IsDir() { if rel, rerr := filepath.Rel(dir, p); rerr == nil { idx = append(idx, rel) } }
        }
        for _, env := range envs {
            glob := filepath.Join(dir, "build", env, "obj", "*", "*.o")
            if matches, _ := filepath.Glob(glob); len(matches) > 0 { for _, m := range matches { if rel, err := filepath.Rel(dir, m); err == nil { objsByEnv[env] = append(objsByEnv[env], rel) } } }
            // per-env objIndex
            for _, e := range ws.Packages {
                ip := filepath.Join(dir, "build", env, "obj", e.Package.Name, "index.json")
                if st, err := os.Stat(ip); err == nil && !st.IsDir() {
                    if rel, rerr := filepath.Rel(dir, ip); rerr == nil {
                        if plan.ObjIndexByEnv == nil { plan.ObjIndexByEnv = map[string][]string{} }
                        plan.ObjIndexByEnv[env] = append(plan.ObjIndexByEnv[env], rel)
                    }
                }
            }
        }
        if len(idx) > 0 { sort.Strings(idx); plan.ObjIndex = append(plan.ObjIndex, idx...) }
        if len(objsByEnv) > 0 { plan.ObjectsByEnv = objsByEnv }
        // top-level objects (non-env)
        var objs []string
        for _, e := range ws.Packages {
            glob := filepath.Join(dir, "build", "obj", e.Package.Name, "*.o")
            if matches, _ := filepath.Glob(glob); len(matches) > 0 {
                for _, m := range matches { if rel, rerr := filepath.Rel(dir, m); rerr == nil { objs = append(objs, rel) } }
            }
        }
        if len(objs) > 0 { sort.Strings(objs); plan.Objects = objs }
        if f, err := os.OpenFile(planPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644); err == nil { _ = json.NewEncoder(f).Encode(plan); _ = f.Close() }
    }

    if jsonOut {
        objects := []string{}
        objectsByEnv := map[string][]string{}
        objIndex := []string{}
        objIndexByEnv := map[string][]string{}
        binaries := []string{}
        binariesByEnv := map[string][]string{}
        for _, e := range ws.Packages {
            glob := filepath.Join(dir, "build", "obj", e.Package.Name, "*.o")
            if matches, _ := filepath.Glob(glob); len(matches) > 0 { objects = append(objects, matches...) }
            idx := filepath.Join(dir, "build", "obj", e.Package.Name, "index.json")
            if st, err := os.Stat(idx); err == nil && !st.IsDir() { objIndex = append(objIndex, idx) }
        }
        _ = filepath.WalkDir(filepath.Join(dir, "build"), func(path string, d os.DirEntry, err error) error {
            if err != nil { return nil }
            if d.IsDir() { b := filepath.Base(path); if b == "debug" || b == "obj" { return filepath.SkipDir }; return nil }
            if info, e := d.Info(); e == nil { mode := info.Mode(); if mode.IsRegular() && (mode&0o111 != 0) { binaries = append(binaries, path) } }
            return nil
        })
        for _, env := range envs {
            glob := filepath.Join(dir, "build", env, "obj", "*", "*.o")
            if matches, _ := filepath.Glob(glob); len(matches) > 0 { objectsByEnv[env] = append(objectsByEnv[env], matches...) }
            _ = filepath.WalkDir(filepath.Join(dir, "build", env), func(path string, d os.DirEntry, err error) error {
                if err != nil { return nil }
                if d.IsDir() { return nil }
                if info, e := d.Info(); e == nil { mode := info.Mode(); if mode.IsRegular() && (mode&0o111 != 0) { binariesByEnv[env] = append(binariesByEnv[env], path) } }
                return nil
            })
            for _, e := range ws.Packages {
                idx := filepath.Join(dir, "build", env, "obj", e.Package.Name, "index.json")
                if st, err := os.Stat(idx); err == nil && !st.IsDir() { objIndexByEnv[env] = append(objIndexByEnv[env], idx) }
            }
        }
        type summary struct {
            Objects       []string            `json:"objects"`
            ObjectIndex   []string            `json:"objIndex"`
            Binaries      []string            `json:"binaries"`
            ObjectsByEnv  map[string][]string `json:"objectsByEnv"`
            ObjIndexByEnv map[string][]string `json:"objIndexByEnv"`
            BinariesByEnv map[string][]string `json:"binariesByEnv"`
            Timestamp     string              `json:"timestamp"`
            Data          map[string]any      `json:"data,omitempty"`
            Code          string              `json:"code,omitempty"`
        }
        rec := summary{
            Objects:       objects,
            ObjectIndex:   objIndex,
            Binaries:      binaries,
            ObjectsByEnv:  objectsByEnv,
            ObjIndexByEnv: objIndexByEnv,
            BinariesByEnv: binariesByEnv,
            Timestamp:     time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
            Data:          map[string]any{"targetDir": absTarget, "targets": envs},
            Code:          "BUILD_OK",
        }
        // include objIndex inside data for verbose summary expectations
        if len(objIndex) > 0 {
            rec.Data["objIndex"] = objIndex
        }
        return json.NewEncoder(out).Encode(rec)
    } else {
        objCount := 0
        for _, e := range ws.Packages { if matches, _ := filepath.Glob(filepath.Join(dir, "build", "obj", e.Package.Name, "*.o")); len(matches) > 0 { objCount += len(matches) } }
        binCount := 0
        var firstBin string
        _ = filepath.WalkDir(filepath.Join(dir, "build"), func(path string, d os.DirEntry, err error) error {
            if err != nil { return nil }
            if d.IsDir() { b := filepath.Base(path); if b == "debug" || b == "obj" { return filepath.SkipDir }; return nil }
            if info, e := d.Info(); e == nil { mode := info.Mode(); if mode.IsRegular() && (mode&0o111 != 0) { if rel, rerr := filepath.Rel(dir, path); rerr == nil { binCount++; if firstBin == "" { firstBin = rel } } } }
            return nil
        })
        if binCount > 0 { fmt.Fprintf(out, "built %d object(s); linked %d binary → %s\n", objCount, binCount, firstBin); return nil }
        fmt.Fprintf(out, "workspace valid: target=%s envs=%d\n", absTarget, len(envs))
        return nil
    }
}
