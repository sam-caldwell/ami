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
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// runBuild validates the workspace and prepares build configuration.
// For this phase, it enforces toolchain.* constraints and emits diagnostics.
func runBuild(out io.Writer, dir string, jsonOut bool) error {
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

    if jsonOut {
        // Emit a simple success summary for consistency with machine parsing.
        rec := diag.Record{
            Timestamp: time.Now().UTC(),
            Level:     diag.Info,
            Code:      "BUILD_OK",
            Message:   "workspace valid; build planning deferred",
            File:      "ami.workspace",
            Data:      map[string]any{"targets": envs, "targetDir": absTarget},
        }
        return json.NewEncoder(out).Encode(rec)
    }
    // Human output kept minimal per SPEC I/O rules.
    fmt.Fprintf(out, "workspace valid: target=%s envs=%d\n", absTarget, len(envs))
    return nil
}
