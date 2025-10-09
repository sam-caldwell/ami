package exec

import (
    "encoding/json"
    "os"
    "path/filepath"
)

// loadWorkersLibFromManifest returns the workers shared library path for a package
// from build/debug/manifest.json when present. The returned path is absolute,
// joined with rootDir.
func loadWorkersLibFromManifest(rootDir, pkg string) string {
    type pkgEntry struct { Name, WorkersLib string }
    var mani struct{ Packages []pkgEntry }
    mb := filepath.Join(rootDir, "build", "debug", "manifest.json")
    if b, err := os.ReadFile(mb); err == nil {
        if err := json.Unmarshal(b, &mani); err == nil {
            for _, p := range mani.Packages {
                if p.Name == pkg && p.WorkersLib != "" {
                    return filepath.Join(rootDir, p.WorkersLib)
                }
            }
        }
    }
    return ""
}

