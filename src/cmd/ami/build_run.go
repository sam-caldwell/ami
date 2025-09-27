package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"

    "github.com/sam-caldwell/ami/src/ami/exit"
    "github.com/sam-caldwell/ami/src/ami/compiler/driver"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
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
    // Consider issues when ami.sum missing or any lists are non-empty as violations for this phase.
    if len(rep.Requirements) > 0 && (!rep.SumFound || len(rep.MissingInSum) > 0 || len(rep.Unsatisfied) > 0 || len(rep.MissingInCache) > 0 || len(rep.Mismatched) > 0 || len(rep.ParseErrors) > 0) {
        if jsonOut {
            rec := diag.Record{
                Timestamp: time.Now().UTC(),
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
            }
            _ = json.NewEncoder(out).Encode(rec)
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
                _, _ = driver.Compile(ws, pkgs, driver.Options{Debug: true})
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
                _, diags := driver.Compile(ws, pkgs, driver.Options{Debug: false})
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
        // Emit a simple success summary for consistency with machine parsing.
        rec := diag.Record{
            Timestamp: time.Now().UTC(),
            Level:     diag.Info,
            Code:      "BUILD_OK",
            Message:   "workspace valid; build planning deferred",
            File:      "ami.workspace",
            Data:      map[string]any{"targets": envs, "targetDir": absTarget, "objIndex": objIdx},
        }
        return json.NewEncoder(out).Encode(rec)
    }
    // Human output kept minimal per SPEC I/O rules.
    fmt.Fprintf(out, "workspace valid: target=%s envs=%d\n", absTarget, len(envs))
    return nil
}
