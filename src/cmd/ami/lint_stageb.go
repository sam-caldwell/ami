package main

import (
    "os"
    "path/filepath"

    "github.com/sam-caldwell/ami/src/ami/compiler/sem"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// lintStageB invokes parser/semantics-backed rules when enabled by toggles t.
// Currently implements MemorySafety diagnostics over all .ami files reachable
// from the main package root and its recursive local imports.
func lintStageB(dir string, ws *workspace.Workspace, t RuleToggles) []diag.Record {
    var out []diag.Record
    if ws == nil { return out }

    // Determine package roots to scan: child-first local imports then main root.
    roots := []string{}
    if p := ws.FindPackage("main"); p != nil && p.Root != "" {
        roots = append(collectLocalImportRoots(ws, p), p.Root)
    }
    if len(roots) == 0 { return out }

    // Gather pragma disables across all roots so we can filter emitted records.
    disables := map[string]map[string]bool{}
    for _, r := range roots {
        m := scanPragmas(dir, r)
        for file, rules := range m { disables[file] = rules }
    }

    // Only run memory safety rules if enabled.
    if !(t.StageB || t.MemorySafety) { return out }

    // Walk each root, analyze every .ami file.
    for _, r := range roots {
        root := filepath.Clean(filepath.Join(dir, r))
        _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
            if err != nil || d.IsDir() || filepath.Ext(path) != ".ami" { return nil }
            // Read file content and create a source.File
            b, err := os.ReadFile(path)
            if err != nil { return nil }
            f := &source.File{Name: path, Content: string(b)}
            // Analyze memory safety and append diagnostics (filter by pragmas)
            ms := sem.AnalyzeMemorySafety(f)
            if len(ms) > 0 {
                for _, d := range ms {
                    if m := disables[path]; m != nil && m[d.Code] { continue }
                    out = append(out, d)
                }
            }
            return nil
        })
    }
    return out
}
