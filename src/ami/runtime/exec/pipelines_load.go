package exec

import (
    "encoding/json"
    "os"
    "path/filepath"
)

// minimal shapes from compiler/driver writePipelinesDebug output
type _pipeList struct {
    Pipelines []_pipeEntry `json:"pipelines"`
}
type _pipeEntry struct {
    Name  string       `json:"name"`
    Steps []_pipeStep  `json:"steps"`
}
type _pipeStep struct {
    Name string   `json:"name"`
    Args []string `json:"args"`
}

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

// loadTransformWorkersFromManifest attempts to use build/debug/manifest.json to find
// unit-specific pipelines.json paths and extract Transform worker names deterministically.
func loadTransformWorkersFromManifest(rootDir, pkg, pipeline string) ([]string, error) {
    type unit struct{ Unit, Pipelines string }
    type pkgEntry struct { Name string; Units []unit }
    var mani struct{ Schema string; Packages []pkgEntry }
    mb := filepath.Join(rootDir, "build", "debug", "manifest.json")
    if b, err := os.ReadFile(mb); err == nil {
        if err := json.Unmarshal(b, &mani); err == nil {
            for _, p := range mani.Packages {
                if p.Name != pkg { continue }
                for _, u := range p.Units {
                    if u.Pipelines == "" { continue }
                    // Pipelines path is emitted relative to project root during compile.
                    pp := filepath.Join(rootDir, u.Pipelines)
                    if workers, ok := extractWorkersFromPipelines(pp, pipeline); ok {
                        return workers, nil
                    }
                }
            }
        }
    }
    return nil, nil
}

func extractWorkersFromPipelines(path, pipeline string) ([]string, bool) {
    type pipeList struct{ Pipelines []struct{ Name string; Steps []struct{ Name string; Args []string } } }
    var pl pipeList
    b, err := os.ReadFile(path)
    if err != nil { return nil, false }
    if err := json.Unmarshal(b, &pl); err != nil { return nil, false }
    for _, pe := range pl.Pipelines {
        if pe.Name != pipeline { continue }
        var ws []string
        for _, s := range pe.Steps { if s.Name == "Transform" && len(s.Args) > 0 { ws = append(ws, s.Args[0]) } }
        return ws, true
    }
    return nil, false
}
