package exec

import (
    "encoding/json"
    "os"
    "path/filepath"
)

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

