package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"

    "github.com/sam-caldwell/ami/src/ami/exit"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// runLint performs basic workspace checks and emits human/JSON output.
func runLint(out io.Writer, dir string, jsonOut bool, verbose bool, strict bool) error {
    wsPath := filepath.Join(dir, "ami.workspace")
    var ws workspace.Workspace
    if _, err := os.Stat(wsPath); err != nil {
        d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_WS_MISSING", Message: "workspace not found", File: "ami.workspace"}
        if jsonOut {
            _ = json.NewEncoder(out).Encode(d)
        } else {
            fmt.Fprintf(out, "lint: %s\n", d.Message)
        }
        return exit.New(exit.User, "%s", d.Message)
    }
    if err := ws.Load(wsPath); err != nil {
        d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_WS_PARSE", Message: fmt.Sprintf("invalid workspace: %v", err), File: "ami.workspace"}
        if jsonOut {
            _ = json.NewEncoder(out).Encode(d)
        } else {
            fmt.Fprintf(out, "lint: %s\n", d.Message)
        }
        return exit.New(exit.User, "%s", d.Message)
    }
    // Basic checks: version non-empty; packages contain main
    if ws.Version == "" || ws.FindPackage("main") == nil {
        d := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Error, Code: "E_WS_SCHEMA", Message: "missing version or main package", File: "ami.workspace"}
        if jsonOut {
            _ = json.NewEncoder(out).Encode(d)
        } else {
            fmt.Fprintf(out, "lint: %s\n", d.Message)
        }
        return exit.New(exit.User, "%s", d.Message)
    }
    // Collect workspace diagnostics.
    var diags []diag.Record
    diags = append(diags, lintWorkspace(dir, &ws)...)
    // Source scaffold: scan for UNKNOWN_IDENT under main package root and recursively in local imports
    if p := ws.FindPackage("main"); p != nil && p.Root != "" {
        roots := append([]string{}, p.Root)
        // child-first ordering for local imports
        roots = append(collectLocalImportRoots(&ws, p), roots...)
        // Deduplicate while preserving order
        seenRoot := map[string]bool{}
        uniq := roots[:0]
        for _, r := range roots {
            if r == "" || seenRoot[r] { continue }
            seenRoot[r] = true
            uniq = append(uniq, r)
        }
        for _, root := range uniq {
            disables := scanPragmas(dir, root)
            srcDiags := scanSourceUnknown(dir, root)
            filtered := srcDiags[:0]
            for _, d := range srcDiags {
                if d.File != "" {
                    if m := disables[d.File]; m != nil && m[d.Code] { continue }
                }
                filtered = append(filtered, d)
            }
            diags = append(diags, filtered...)
        }
    }

    // Stage B placeholder: parser/semantics-backed rules (no-op until frontend is integrated)
    if currentRuleToggles.StageB || currentRuleToggles.UnknownIdent || currentRuleToggles.Unused || currentRuleToggles.ImportExist || currentRuleToggles.Duplicates || currentRuleToggles.MemorySafety || currentRuleToggles.RAIIHint {
        diags = append(diags, lintStageB(dir, &ws, currentRuleToggles)...)
    }

    // Check workspace config for strict if not provided via flag.
    if !strict {
        for _, opt := range ws.Toolchain.Linter.Options {
            if opt == "strict" { strict = true; break }
        }
    }

    // Apply severity mapping from config rules
    if m := ws.Toolchain.Linter.Rules; len(m) > 0 {
        mapped := make([]diag.Record, 0, len(diags))
        for _, d := range diags {
            if lvl, ok := m[d.Code]; ok {
                switch lvl {
                case "off":
                    continue
                case "info":
                    d.Level = diag.Info
                case "warn":
                    d.Level = diag.Warn
                case "error":
                    d.Level = diag.Error
                }
            }
            mapped = append(mapped, d)
        }
        diags = mapped
    }

    // Summarize
    var errorsN, warnsN int
    for _, d := range diags {
        if d.Level == diag.Error { errorsN++ }
        if d.Level == diag.Warn { warnsN++ }
    }
    // Strict: promote warnings to errors
    if strict && warnsN > 0 {
        for i := range diags {
            if diags[i].Level == diag.Warn { diags[i].Level = diag.Error }
        }
        errorsN += warnsN
        warnsN = 0
    }
    // Optional debug stream to file when verbose
    var debugFile io.WriteCloser
    if verbose {
        debugPath := filepath.Join(dir, "build", "debug")
        _ = os.MkdirAll(debugPath, 0o755)
        f, err := os.OpenFile(filepath.Join(debugPath, "lint.ndjson"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
        if err == nil { debugFile = f }
    }
    // Helper to optionally write diag to debug file as NDJSON
    writeDebug := func(d diag.Record) {
        if debugFile != nil { _, _ = debugFile.Write(diag.Line(d)) }
    }
    if jsonOut {
        enc := json.NewEncoder(out)
        for _, d := range diags { _ = enc.Encode(d); writeDebug(d) }
        // final summary record
        sum := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Info, Code: "SUMMARY", Message: "lint summary", Data: map[string]any{"errors": errorsN, "warnings": warnsN}}
        writeDebug(sum)
        _ = enc.Encode(sum)
        if debugFile != nil { _ = debugFile.Close() }
        if errorsN > 0 { return exit.New(exit.User, "%s", "lint errors found") }
        return nil
    }
    if len(diags) == 0 {
        fmt.Fprintln(out, "lint: OK")
        if debugFile != nil { _ = debugFile.Close() }
        return nil
    }
    for _, d := range diags {
        if d.Pos != nil {
            fmt.Fprintf(out, "lint: %s %s: %s (%s:%d:%d)\n", string(d.Level), d.Code, d.Message, d.File, d.Pos.Line, d.Pos.Column)
        } else {
            fmt.Fprintf(out, "lint: %s %s: %s (%s)\n", string(d.Level), d.Code, d.Message, d.File)
        }
        writeDebug(d)
    }
    sum := diag.Record{Timestamp: time.Now().UTC(), Level: diag.Info, Code: "SUMMARY", Message: "lint summary", Data: map[string]any{"errors": errorsN, "warnings": warnsN}}
    writeDebug(sum)
    if debugFile != nil { _ = debugFile.Close() }
    fmt.Fprintf(out, "lint: %d error(s), %d warning(s)\n", errorsN, warnsN)
    if errorsN > 0 { return exit.New(exit.User, "%s", "lint errors found") }
    return nil
}
