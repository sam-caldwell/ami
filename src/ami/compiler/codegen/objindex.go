package codegen

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "io/fs"
    "os"
    "path/filepath"
    "sort"
)

// ObjIndex is the on-disk schema for build/obj/<package>/index.json (objindex.v1).
type ObjIndex struct {
    Schema string       `json:"schema"`
    Package string      `json:"package"`
    Units  []ObjUnit    `json:"units"`
}

// ObjUnit describes a single object unit entry.
type ObjUnit struct {
    Unit  string `json:"unit"`
    Path  string `json:"path"`
    Size  int64  `json:"size"`
    Sha256 string `json:"sha256"`
}

// BuildObjIndex scans the directory and produces a deterministic index.
// It indexes files with extension .s (asm) and .o (object stubs).
func BuildObjIndex(pkg string, dir string) (ObjIndex, error) {
    var files []string
    err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
        if err != nil { return err }
        if d.IsDir() { return nil }
        ext := filepath.Ext(p)
        if ext != ".s" && ext != ".o" { return nil }
        rel, err := filepath.Rel(dir, p)
        if err != nil { return err }
        files = append(files, rel)
        return nil
    })
    if err != nil { return ObjIndex{}, err }
    sort.Strings(files)
    units := make([]ObjUnit, 0, len(files))
    for _, rel := range files {
        p := filepath.Join(dir, rel)
        st, err := os.Stat(p)
        if err != nil { return ObjIndex{}, err }
        b, err := os.ReadFile(p)
        if err != nil { return ObjIndex{}, err }
        h := sha256.Sum256(b)
        units = append(units, ObjUnit{
            Unit:   trimExt(rel),
            Path:   rel,
            Size:   st.Size(),
            Sha256: hex.EncodeToString(h[:]),
        })
    }
    return ObjIndex{Schema: "objindex.v1", Package: pkg, Units: units}, nil
}

// WriteObjIndex writes the index JSON to build/obj/<package>/index.json.
func WriteObjIndex(idx ObjIndex) error {
    base := filepath.Join("build", "obj", idx.Package)
    if err := os.MkdirAll(base, 0o755); err != nil { return err }
    b, err := json.MarshalIndent(idx, "", "  ")
    if err != nil { return err }
    return os.WriteFile(filepath.Join(base, "index.json"), b, 0o644)
}

func trimExt(path string) string {
    ext := filepath.Ext(path)
    return path[:len(path)-len(ext)]
}
