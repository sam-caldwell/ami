package exec

import (
    "encoding/json"
    "os"
    "path/filepath"
)

// loadTransformWorkers returns the list of worker names for Transform steps
// in order of appearance for the named pipeline within package `pkg`.
// It searches build/debug/ir/<pkg>/*.pipelines.json under rootDir and returns
// the first matching pipeline found.
func loadTransformWorkers(rootDir, pkg, pipeline string) ([]string, error) {
    // Prefer build/debug/manifest.json when present for deterministic unit selection.
    if ws, err := loadTransformWorkersFromManifest(rootDir, pkg, pipeline); err == nil && ws != nil {
        return ws, nil
    }
    dir := filepath.Join(rootDir, "build", "debug", "ir", pkg)
    entries, err := os.ReadDir(dir)
    if err != nil { return nil, err }
    for _, e := range entries {
        if e.IsDir() { continue }
        name := e.Name()
        if len(name) < len(".pipelines.json") || name[len(name)-len(".pipelines.json"):] != ".pipelines.json" { continue }
        b, err := os.ReadFile(filepath.Join(dir, name))
        if err != nil { continue }
        var pl _pipeList
        if err := json.Unmarshal(b, &pl); err != nil { continue }
        for _, pe := range pl.Pipelines {
            if pe.Name != pipeline { continue }
            var workers []string
            for _, s := range pe.Steps {
                if s.Name != "Transform" { continue }
                if len(s.Args) == 0 { continue }
                wname := s.Args[0]
                // Trim quotes if present; JSON args come from parser.Text, may include quotes
                if len(wname) >= 2 && ((wname[0] == '"' && wname[len(wname)-1] == '"') || (wname[0] == '\'' && wname[len(wname)-1] == '\'')) {
                    wname = wname[1:len(wname)-1]
                }
                workers = append(workers, wname)
            }
            return workers, nil
        }
    }
    return nil, nil
}

