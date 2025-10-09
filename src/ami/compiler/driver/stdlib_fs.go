package driver

import (
    "os"
    "path/filepath"
    "sort"
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// fsStdlibPackages loads stdlib packages from the given root directory.
// It expects structure: <root>/<pkg>/*.ami. Returns packages sorted by name.
func fsStdlibPackages(root string) []Package {
    var out []Package
    entries, err := os.ReadDir(root)
    if err != nil { return nil }
    for _, e := range entries {
        if !e.IsDir() { continue }
        pkgName := e.Name()
        dir := filepath.Join(root, pkgName)
        files, _ := os.ReadDir(dir)
        fs := &source.FileSet{}
        for _, f := range files {
            if f.IsDir() { continue }
            name := f.Name()
            if !strings.HasSuffix(name, ".ami") { continue }
            b, err := os.ReadFile(filepath.Join(dir, name))
            if err != nil { continue }
            fs.AddFile(name, string(b))
        }
        if len(fs.Files) == 0 { continue }
        out = append(out, Package{Name: pkgName, Files: fs})
    }
    sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
    return out
}

