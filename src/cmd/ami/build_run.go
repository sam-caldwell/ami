package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/exit"
    "github.com/sam-caldwell/ami/src/ami/compiler/driver"
    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

func containsEnv(list []string, env string) bool {
    for _, e := range list { if e == env { return true } }
    return false
}

// linkExtraFlags returns a set of linker flags adjusted for the target env
// and workspace linker options.
func linkExtraFlags(env string, opts []string) []string {
    var extra []string
    // Default: on Darwin, prefer dead strip
    if env == "darwin/arm64" || env == "darwin/amd64" || env == "darwin/x86_64" {
        extra = append(extra, "-Wl,-dead_strip")
    }
    // Options mapping
    for _, opt := range opts {
        switch opt {
        case "PIE", "pie":
            if env == "darwin/arm64" || env == "darwin/amd64" || env == "darwin/x86_64" {
                extra = append(extra, "-Wl,-pie")
            } else {
                extra = append(extra, "-pie")
            }
        case "static":
            // Best effort: static commonly supported on Linux
            if strings.HasPrefix(env, "linux/") { extra = append(extra, "-static") }
        case "dead_strip", "dce":
            if strings.HasPrefix(env, "darwin/") { extra = append(extra, "-Wl,-dead_strip") }
            if strings.HasPrefix(env, "linux/") { extra = append(extra, "-Wl,--gc-sections") }
        }
    }
    return extra
}

// runBuild validates the workspace and prepares build configuration.
// For this phase, it enforces toolchain.* constraints and emits diagnostics.
func runBuild(out io.Writer, dir string, jsonOut bool, verbose bool) error {
    // Thin wrapper to keep primary logic isolated for readability.
    return runBuildImpl(out, dir, jsonOut, verbose)
}

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

    // If ami.manifest exists alongside ami.sum, perform a simple presence check:
    // compare name@version pairs irrespective of sha. Mismatch yields E_INTEGRITY_MANIFEST.
    maniPath := filepath.Join(dir, "ami.manifest")
    if st, err := os.Stat(maniPath); err == nil && !st.IsDir() {
        var sum workspace.Manifest
        var mani workspace.Manifest
        sumPath := filepath.Join(dir, "ami.sum")
        if st, err := os.Stat(sumPath); err == nil && !st.IsDir() {
            _ = sum.Load(sumPath)
        }
        _ = mani.Load(maniPath)
        // Build comparable sets
        type pair struct{ n, v string }
        have := map[pair]bool{}
        want := map[pair]bool{}
        for name, vers := range sum.Packages { for ver := range vers { have[pair{name, ver}] = true } }
        for name, vers := range mani.Packages { for ver := range vers { want[pair{name, ver}] = true } }
        // detect any difference
        mismatch := false
        if len(have) != len(want) { mismatch = true }
        if !mismatch {
            for k := range have { if !want[k] { mismatch = true; break } }
            if !mismatch { for k := range want { if !have[k] { mismatch = true; break } } }
        }
        if mismatch {
            if jsonOut {
                _ = json.NewEncoder(out).Encode(diag.Record{
                    Timestamp: time.Now().UTC(),
                    Level:     diag.Error,
                    Code:      "E_INTEGRITY_MANIFEST",
                    Message:   "ami.manifest disagrees with ami.sum",
                    File:      "ami.manifest",
                    Data: map[string]any{
                        "sum":  sum.Packages,
                        "mani": mani.Packages,
                    },
                })
            }
            return exit.New(exit.Integrity, "manifest mismatch")
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
        // Front-end debug artifacts: parse all workspace packages and compile with Debug=true
        var dbgPkgs []driver.Package
        for _, entry := range ws.Packages {
            p := entry.Package
            if p.Root == "" { continue }
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
            if len(files) == 0 { continue }
            var fs source.FileSet
            for _, f := range files {
                b, err := os.ReadFile(f)
                if err != nil { continue }
                fs.AddFile(f, string(b))
            }
            dbgPkgs = append(dbgPkgs, driver.Package{Name: p.Name, Files: &fs})
        }
        if len(dbgPkgs) > 0 {
            // Run compile with CWD set to workspace dir so relative debug paths land under dir/build/debug
            oldwd, _ := os.Getwd()
            _ = os.Chdir(dir)
            // hook logger for full timestamped compiler activity under build/debug/activity.log
            var logcb func(string, map[string]any)
            if lg := getRootLogger(); lg != nil {
                logcb = func(event string, fields map[string]any) { lg.Info("compiler."+event, fields) }
                // Record toolchain versions (clang) for visibility in verbose logs
                if clang, err := llvme.FindClang(); err == nil {
                    if ver, verr := llvme.Version(clang); verr == nil { lg.Info("toolchain.clang", map[string]any{"version": ver, "path": clang}) }
                } else { lg.Info("toolchain.missing", map[string]any{"tool": "clang"}) }
            }
            _, _ = driver.Compile(ws, dbgPkgs, driver.Options{Debug: true, EmitLLVMOnly: buildEmitLLVMOnly, NoLink: buildNoLink, Log: logcb})
            _ = os.Chdir(oldwd)
        }
        // Build plan after emitting artifacts; include object index paths when present
        planPath := filepath.Join(dir, "build", "debug", "build.plan.json")
        type planPkg struct{
            Key, Name, Version, Root string
            HasObjects bool `json:"hasObjects"`
        }
        plan := struct {
            Schema    string    `json:"schema"`
            TargetDir string    `json:"targetDir"`
            Targets   []string  `json:"targets"`
            Packages  []planPkg `json:"packages"`
            ObjIndex  []string  `json:"objIndex,omitempty"`
            Objects   []string  `json:"objects,omitempty"`
            ObjectsByEnv map[string][]string `json:"objectsByEnv,omitempty"`
            ObjIndexByEnv map[string][]string `json:"objIndexByEnv,omitempty"`
        }{Schema: "build.plan/v1", TargetDir: absTarget, Targets: envs}
        for _, e := range ws.Packages {
            // detect any object files for this package
            hasObjects := false
            if matches, _ := filepath.Glob(filepath.Join(dir, "build", "obj", e.Package.Name, "*.o")); len(matches) > 0 { hasObjects = true }
            plan.Packages = append(plan.Packages, planPkg{Key: e.Key, Name: e.Package.Name, Version: e.Package.Version, Root: e.Package.Root, HasObjects: hasObjects})
            // if object index exists for this package, include path
            idx := filepath.Join(dir, "build", "obj", e.Package.Name, "index.json")
            if st, err := os.Stat(idx); err == nil && !st.IsDir() {
                rel, _ := filepath.Rel(dir, idx)
                plan.ObjIndex = append(plan.ObjIndex, rel)
            }
            // Include .o object files when present
            glob := filepath.Join(dir, "build", "obj", e.Package.Name, "*.o")
            if matches, _ := filepath.Glob(glob); len(matches) > 0 {
                for _, m := range matches { if rel, err := filepath.Rel(dir, m); err == nil { plan.Objects = append(plan.Objects, rel) } }
            }
        }
        // Include per-env objects if present under build/<env>/obj/**
        if len(envs) > 0 {
            plan.ObjectsByEnv = map[string][]string{}
            plan.ObjIndexByEnv = map[string][]string{}
            for _, env := range envs {
                var list []string
                for _, e := range ws.Packages {
                    glob := filepath.Join(dir, "build", env, "obj", e.Package.Name, "*.o")
                    if matches, _ := filepath.Glob(glob); len(matches) > 0 {
                        for _, m := range matches { if rel, err := filepath.Rel(dir, m); err == nil { list = append(list, rel) } }
                    }
                    // per-env obj index file
                    idx := filepath.Join(dir, "build", env, "obj", e.Package.Name, "index.json")
                    if st, err := os.Stat(idx); err == nil && !st.IsDir() {
                        if rel, rerr := filepath.Rel(dir, idx); rerr == nil { plan.ObjIndexByEnv[env] = append(plan.ObjIndexByEnv[env], rel) }
                    }
                }
                if len(list) > 0 { sort.Strings(list); plan.ObjectsByEnv[env] = list }
            }
            if len(plan.ObjectsByEnv) == 0 { plan.ObjectsByEnv = nil }
            if len(plan.ObjIndexByEnv) == 0 { plan.ObjIndexByEnv = nil }
        }
        if f, err := os.OpenFile(planPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644); err == nil {
            _ = json.NewEncoder(f).Encode(plan)
            _ = f.Close()
        }
    }

    // In JSON mode, run a non-debug compile to surface parser/semantic diagnostics as a stream.
    if jsonOut {
        // Compile all packages (non-debug) to surface diagnostics as a stream
        var pkgs []driver.Package
        for _, entry := range ws.Packages {
            p := entry.Package
            if p.Root == "" { continue }
            root := filepath.Clean(filepath.Join(dir, p.Root))
            var files []string
            _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
                if err != nil || d.IsDir() { return nil }
                if filepath.Ext(path) == ".ami" { files = append(files, path) }
                return nil
            })
            if len(files) == 0 { continue }
            var fs source.FileSet
            for _, f := range files { b, err := os.ReadFile(f); if err == nil { fs.AddFile(f, string(b)) } }
            pkgs = append(pkgs, driver.Package{Name: p.Name, Files: &fs})
        }
        if len(pkgs) > 0 {
            var logcb func(string, map[string]any)
            if lg := getRootLogger(); lg != nil {
                logcb = func(event string, fields map[string]any) { lg.Info("compiler."+event, fields) }
            }
            oldwd, _ := os.Getwd()
            _ = os.Chdir(dir)
            _, diags := driver.Compile(ws, pkgs, driver.Options{Debug: false, EmitLLVMOnly: buildEmitLLVMOnly, NoLink: buildNoLink, Log: logcb})
            _ = os.Chdir(oldwd)
            if len(diags) > 0 {
                enc := json.NewEncoder(out)
                for i := range diags { _ = enc.Encode(diags[i]) }
                return exit.New(exit.User, "compiler reported diagnostics")
            }
        }
    }

    // Always perform a non-debug compile pass to emit object stubs + object index under build/obj for all packages
    {
        var pkgs []driver.Package
        for _, entry := range ws.Packages {
            p := entry.Package
            if p.Root == "" { continue }
            root := filepath.Clean(filepath.Join(dir, p.Root))
            var files []string
            _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
                if err != nil || d.IsDir() { return nil }
                if filepath.Ext(path) == ".ami" { files = append(files, path) }
                return nil
            })
            if len(files) == 0 { continue }
            var fs source.FileSet
            for _, f := range files { b, err := os.ReadFile(f); if err == nil { fs.AddFile(f, string(b)) } }
            pkgs = append(pkgs, driver.Package{Name: p.Name, Files: &fs})
        }
        if len(pkgs) > 0 {
            // Run compile with CWD set to workspace dir so outputs land under dir/build
            oldwd, _ := os.Getwd()
            _ = os.Chdir(dir)
            _, _ = driver.Compile(ws, pkgs, driver.Options{Debug: false, EmitLLVMOnly: buildEmitLLVMOnly, NoLink: buildNoLink})
            _ = os.Chdir(oldwd)
        }
    }

    // Link stage — produce executables per-env when possible, and fall back to default objects when no per-env objects are present.
    if !buildNoLink {
        clang, ferr := llvme.FindClang()
        if ferr != nil {
            if lg := getRootLogger(); lg != nil { lg.Info("build.toolchain.missing", map[string]any{"tool": "clang"}) }
        } else {
            // Resolve binary name from main package
            binName := "app"
            if mp := ws.FindPackage("main"); mp != nil && mp.Name != "" { binName = mp.Name }

            // First, attempt per-env linking where env-specific objects exist.
            envWithObjects := map[string]bool{}
            for _, env := range envs {
                if containsEnv(buildNoLinkEnvs, env) { continue }
                // collect per-env objects
                var objs []string
                for _, e := range ws.Packages {
                    glob := filepath.Join(dir, "build", env, "obj", e.Package.Name, "*.o")
                    if matches, _ := filepath.Glob(glob); len(matches) > 0 { objs = append(objs, matches...) }
                }
                if len(objs) == 0 { continue }
                envWithObjects[env] = true
                triple := llvme.TripleForEnv(env)
                rtDir := filepath.Join(dir, "build", "runtime")
                if llPath, werr := llvme.WriteRuntimeLL(rtDir, triple, true); werr == nil {
                    rtObj := filepath.Join(rtDir, "runtime.o")
                    if cerr := llvme.CompileLLToObject(clang, llPath, rtObj, triple); cerr == nil {
                        objs = append(objs, rtObj)
                        outDir := filepath.Join(dir, "build", env)
                        _ = os.MkdirAll(outDir, 0o755)
                        outBin := filepath.Join(outDir, binName)
                        extra := linkExtraFlags(env, ws.Toolchain.Linker.Options)
                        if lerr := llvme.LinkObjects(clang, objs, outBin, triple, extra...); lerr != nil {
                            if lg := getRootLogger(); lg != nil { lg.Info("build.link.fail", map[string]any{"error": lerr.Error(), "bin": outBin, "env": env}) }
                            if jsonOut {
                                d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_LINK_FAIL", Message: "linking failed", File: "clang", Data: map[string]any{"env": env}}
                                if te, ok := lerr.(llvme.ToolError); ok { d.Data["stderr"] = te.Stderr }
                                _ = json.NewEncoder(out).Encode(d)
                            }
                        } else if lg := getRootLogger(); lg != nil {
                            lg.Info("build.link.ok", map[string]any{"bin": outBin, "objects": len(objs), "env": env})
                        }
                    } else if lg := getRootLogger(); lg != nil {
                        lg.Info("build.runtime.obj.fail", map[string]any{"error": cerr.Error(), "env": env})
                    }
                } else if lg := getRootLogger(); lg != nil {
                    lg.Info("build.runtime.ll.fail", map[string]any{"error": werr.Error(), "env": env})
                }
            }

            // Fallback: link default objects under build/obj when no per-env objects exist for the primary env.
            // This preserves existing behavior.
            if len(envs) == 0 || !envWithObjects[envs[0]] {
                if len(envs) > 0 && containsEnv(buildNoLinkEnvs, envs[0]) { /* skip fallback link */ } else {
                var objects []string
                for _, e := range ws.Packages {
                    glob := filepath.Join(dir, "build", "obj", e.Package.Name, "*.o")
                    if matches, _ := filepath.Glob(glob); len(matches) > 0 { objects = append(objects, matches...) }
                }
                if len(objects) > 0 {
                    triple := llvme.DefaultTriple
                    if len(envs) > 0 { triple = llvme.TripleForEnv(envs[0]) }
                    rtDir := filepath.Join(dir, "build", "runtime")
                    if llPath, werr := llvme.WriteRuntimeLL(rtDir, triple, true); werr == nil {
                        rtObj := filepath.Join(rtDir, "runtime.o")
                        if cerr := llvme.CompileLLToObject(clang, llPath, rtObj, triple); cerr == nil {
                            objects = append(objects, rtObj)
                            outDir := filepath.Join(dir, "build")
                            if len(envs) > 0 && envs[0] != "" { outDir = filepath.Join(outDir, envs[0]) }
                            _ = os.MkdirAll(outDir, 0o755)
                            outBin := filepath.Join(outDir, binName)
                            env0 := ""
                            if len(envs) > 0 { env0 = envs[0] }
                            extra := linkExtraFlags(env0, ws.Toolchain.Linker.Options)
                            if lerr := llvme.LinkObjects(clang, objects, outBin, triple, extra...); lerr != nil {
                                if lg := getRootLogger(); lg != nil { lg.Info("build.link.fail", map[string]any{"error": lerr.Error(), "bin": outBin}) }
                                if jsonOut {
                                    d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_LINK_FAIL", Message: "linking failed", File: "clang"}
                                    if te, ok := lerr.(llvme.ToolError); ok { if d.Data == nil { d.Data = map[string]any{} }; d.Data["stderr"] = te.Stderr }
                                    _ = json.NewEncoder(out).Encode(d)
                                }
                            } else if lg := getRootLogger(); lg != nil {
                                lg.Info("build.link.ok", map[string]any{"bin": outBin, "objects": len(objects)})
                            }
                        } else if lg := getRootLogger(); lg != nil {
                            lg.Info("build.runtime.obj.fail", map[string]any{"error": cerr.Error()})
                        }
                    } else if lg := getRootLogger(); lg != nil {
                        lg.Info("build.runtime.ll.fail", map[string]any{"error": werr.Error()})
                    }
                }
            }
        }
    }
    }
    // Close link stage block explicitly before manifest rewrite.
    // (Ensures subsequent steps are outside of the link conditional.)

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
        // Include objects when present for visibility + structured artifacts with kind:"obj"
        var objects []string
        var artifacts []map[string]any
        for _, e := range ws.Packages {
            glob := filepath.Join(dir, "build", "obj", e.Package.Name, "*.o")
            if matches, _ := filepath.Glob(glob); len(matches) > 0 {
                for _, m := range matches {
                    if rel, err := filepath.Rel(dir, m); err == nil {
                        objects = append(objects, rel)
                        artifacts = append(artifacts, map[string]any{"kind": "obj", "path": rel})
                    }
                }
            }
        }
        if len(objects) > 0 { sort.Strings(objects); outObj["objects"] = objects }
        if len(artifacts) > 0 { outObj["artifacts"] = artifacts }
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
        // Per-env binaries
        binsByEnv := map[string][]string{}
        for _, env := range envs {
            envDir := filepath.Join(buildDir, env)
            var list []string
            _ = filepath.WalkDir(envDir, func(path string, d os.DirEntry, err error) error {
                if err != nil { return nil }
                if d.IsDir() { return nil }
                if info, e := d.Info(); e == nil {
                    mode := info.Mode()
                    if mode.IsRegular() && (mode&0o111 != 0) {
                        if rel, rerr := filepath.Rel(dir, path); rerr == nil { list = append(list, rel) }
                    }
                }
                return nil
            })
            if len(list) > 0 { sort.Strings(list); binsByEnv[env] = list }
        }
        if len(binsByEnv) > 0 { outObj["binariesByEnv"] = binsByEnv }
        // Include per-env summaries if present
        objectsByEnv := map[string][]string{}
        objIndexByEnv := map[string][]string{}
        for _, env := range envs {
            var objs []string
            for _, e := range ws.Packages {
                glob := filepath.Join(dir, "build", env, "obj", e.Package.Name, "*.o")
                if matches, _ := filepath.Glob(glob); len(matches) > 0 {
                    for _, m := range matches { if rel, err := filepath.Rel(dir, m); err == nil { objs = append(objs, rel) } }
                }
                idx := filepath.Join(dir, "build", env, "obj", e.Package.Name, "index.json")
                if st, err := os.Stat(idx); err == nil && !st.IsDir() {
                    if rel, rerr := filepath.Rel(dir, idx); rerr == nil { objIndexByEnv[env] = append(objIndexByEnv[env], rel) }
                }
            }
            if len(objs) > 0 { sort.Strings(objs); objectsByEnv[env] = objs }
        }
        if len(objectsByEnv) > 0 { outObj["objectsByEnv"] = objectsByEnv }
        if len(objIndexByEnv) > 0 { outObj["objIndexByEnv"] = objIndexByEnv }
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
        // Collect object index paths when available
        var objIdx []string
        for _, e := range ws.Packages {
            idx := filepath.Join(dir, "build", "obj", e.Package.Name, "index.json")
            if st, err := os.Stat(idx); err == nil && !st.IsDir() {
                if rel, rerr := filepath.Rel(dir, idx); rerr == nil { objIdx = append(objIdx, rel) }
            }
        }
        sort.Strings(objIdx)
        // Discover binaries under build/ (exclude debug/ and obj/)
        var bins []string
        _ = filepath.WalkDir(filepath.Join(dir, "build"), func(path string, d os.DirEntry, err error) error {
            if err != nil { return nil }
            if d.IsDir() {
                b := filepath.Base(path)
                if b == "debug" || b == "obj" { return filepath.SkipDir }
                return nil
            }
            if info, e := d.Info(); e == nil {
                mode := info.Mode()
                if mode.IsRegular() && (mode&0o111 != 0) {
                    if rel, rerr := filepath.Rel(dir, path); rerr == nil { bins = append(bins, rel) }
                }
            }
            return nil
        })
        if len(bins) > 0 { sort.Strings(bins) }
        // Per-env summaries
        objectsByEnv := map[string][]string{}
        objIndexByEnv := map[string][]string{}
        binariesByEnv := map[string][]string{}
        for _, env := range envs {
            // objects
            var objs []string
            for _, e := range ws.Packages {
                glob := filepath.Join(dir, "build", env, "obj", e.Package.Name, "*.o")
                if matches, _ := filepath.Glob(glob); len(matches) > 0 {
                    for _, m := range matches { if rel, err := filepath.Rel(dir, m); err == nil { objs = append(objs, rel) } }
                }
                idx := filepath.Join(dir, "build", env, "obj", e.Package.Name, "index.json")
                if st, err := os.Stat(idx); err == nil && !st.IsDir() {
                    if rel, rerr := filepath.Rel(dir, idx); rerr == nil { objIndexByEnv[env] = append(objIndexByEnv[env], rel) }
                }
            }
            if len(objs) > 0 { sort.Strings(objs); objectsByEnv[env] = objs }
            // binaries
            envDir := filepath.Join(dir, "build", env)
            var blist []string
            _ = filepath.WalkDir(envDir, func(path string, d os.DirEntry, err error) error {
                if err != nil { return nil }
                if d.IsDir() { return nil }
                if info, e := d.Info(); e == nil {
                    mode := info.Mode()
                    if mode.IsRegular() && (mode&0o111 != 0) {
                        if rel, rerr := filepath.Rel(dir, path); rerr == nil { blist = append(blist, rel) }
                    }
                }
                return nil
            })
            if len(blist) > 0 { sort.Strings(blist); binariesByEnv[env] = blist }
        }
        if len(objectsByEnv) == 0 { objectsByEnv = nil }
        if len(objIndexByEnv) == 0 { objIndexByEnv = nil }
        if len(binariesByEnv) == 0 { binariesByEnv = nil }
        rec := diag.Record{
            Timestamp: time.Now().UTC(),
            Level:     diag.Info,
            Code:      "BUILD_OK",
            Message:   "workspace valid; build planning deferred",
            File:      "ami.workspace",
            Data: map[string]any{
                "targets":       envs,
                "targetDir":     absTarget,
                "objIndex":      objIdx,
                "buildManifest": filepath.Join("build", "ami.manifest"),
                "binaries":      bins,
                "objectsByEnv":  objectsByEnv,
                "objIndexByEnv": objIndexByEnv,
                "binariesByEnv": binariesByEnv,
            },
        }
        return json.NewEncoder(out).Encode(rec)
    } else {
        // Human summary
        objCount := 0
        for _, e := range ws.Packages {
            glob := filepath.Join(dir, "build", "obj", e.Package.Name, "*.o")
            if matches, _ := filepath.Glob(glob); len(matches) > 0 { objCount += len(matches) }
        }
        binCount := 0
        var firstBin string
        _ = filepath.WalkDir(filepath.Join(dir, "build"), func(path string, d os.DirEntry, err error) error {
            if err != nil { return nil }
            if d.IsDir() {
                b := filepath.Base(path)
                if b == "debug" || b == "obj" { return filepath.SkipDir }
                return nil
            }
            if info, e := d.Info(); e == nil {
                mode := info.Mode()
                if mode.IsRegular() && (mode&0o111 != 0) {
                    if rel, rerr := filepath.Rel(dir, path); rerr == nil {
                        binCount++
                        if firstBin == "" { firstBin = rel }
                    }
                }
            }
            return nil
        })
        if binCount > 0 {
            fmt.Fprintf(out, "built %d object(s); linked %d binary → %s\n", objCount, binCount, firstBin)
            return nil
        }
        fmt.Fprintf(out, "workspace valid: target=%s envs=%d\n", absTarget, len(envs))
        return nil
    }
}
