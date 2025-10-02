package codegen

import (
    "crypto/sha256"
    "encoding/hex"
    "io/fs"
    "os"
    "path/filepath"
    "sort"
)

// BuildObjIndex scans the directory and produces a deterministic index.
// It indexes files with extension .s (asm) and .o (object). When both exist
// for the same unit, .o is preferred and only one entry is emitted.
func BuildObjIndex(pkg string, dir string) (ObjIndex, error) {
    // unit -> chosen relative path (prefer .o over .s)
    chosen := map[string]string{}
    err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
        if err != nil { return err }
        if d.IsDir() { return nil }
        ext := filepath.Ext(p)
        if ext != ".s" && ext != ".o" { return nil }
        rel, err := filepath.Rel(dir, p)
        if err != nil { return err }
        unit := trimExt(rel)
        if prev, ok := chosen[unit]; ok {
            // If we already chose .s and we found .o, upgrade; else keep previous
            if filepath.Ext(prev) == ".s" && ext == ".o" { chosen[unit] = rel }
            return nil
        }
        chosen[unit] = rel
        return nil
    })
    if err != nil { return ObjIndex{}, err }
    // Deterministic unit ordering
    var unitsOrdered []string
    for u := range chosen { unitsOrdered = append(unitsOrdered, u) }
    sort.Strings(unitsOrdered)
    units := make([]ObjUnit, 0, len(unitsOrdered))
    for _, unit := range unitsOrdered {
        rel := chosen[unit]
        p := filepath.Join(dir, rel)
        st, err := os.Stat(p)
        if err != nil { return ObjIndex{}, err }
        b, err := os.ReadFile(p)
        if err != nil { return ObjIndex{}, err }
        h := sha256.Sum256(b)
        units = append(units, ObjUnit{
            Unit:   unit,
            Path:   rel,
            Size:   st.Size(),
            Sha256: hex.EncodeToString(h[:]),
        })
    }
    return ObjIndex{Schema: "objindex.v1", Package: pkg, Units: units}, nil
}

