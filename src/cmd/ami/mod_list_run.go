package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "regexp"
    "sort"
    "time"
)

// types moved to mod_list_entry.go and mod_list_result.go

func runModList(out io.Writer, jsonOut bool) error {
    cache := os.Getenv("AMI_PACKAGE_CACHE")
    if cache == "" {
        home, _ := os.UserHomeDir()
        cache = filepath.Join(home, ".ami", "pkg")
    }
    _ = os.MkdirAll(cache, 0o755)
    dirents, _ := os.ReadDir(cache)
    entries := []modListEntry{}
    // Simple SemVer detection: v?X.Y.Z with optional prerelease
    semver := regexp.MustCompile(`^[vV]?\d+\.\d+\.\d+(?:[-.][0-9A-Za-z.-]+)?$`)
    for _, e := range dirents {
        info, err := e.Info()
        if err != nil { continue }
        if e.IsDir() {
            base := filepath.Join(cache, e.Name())
            kids, _ := os.ReadDir(base)
            versioned := false
            for _, k := range kids {
                if k.IsDir() && semver.MatchString(k.Name()) {
                    ki, err := k.Info()
                    if err != nil { continue }
                    entries = append(entries, modListEntry{
                        Name:     e.Name(),
                        Version:  k.Name(),
                        Type:     "dir",
                        Size:     ki.Size(),
                        Modified: ki.ModTime().UTC().Format(time.RFC3339Nano),
                    })
                    versioned = true
                }
            }
            if versioned { continue }
        }
        t := "file"
        if e.IsDir() { t = "dir" }
        entries = append(entries, modListEntry{
            Name:     e.Name(),
            Type:     t,
            Size:     info.Size(),
            Modified: info.ModTime().UTC().Format(time.RFC3339Nano),
        })
    }
    sort.Slice(entries, func(i, j int) bool {
        if entries[i].Name == entries[j].Name {
            return entries[i].Version < entries[j].Version
        }
        return entries[i].Name < entries[j].Name
    })
    res := modListResult{Path: cache, Entries: entries}
    if jsonOut {
        return json.NewEncoder(out).Encode(res)
    }
    for _, e := range entries {
        display := e.Name
        if e.Version != "" { display = e.Name + "@" + e.Version }
        _, _ = fmt.Fprintf(out, "%s\t%s\t%d\t%s\n", e.Type, display, e.Size, e.Modified)
    }
    return nil
}
