package main

import (
    "bufio"
    "os"
    "path/filepath"
    "strings"
)

// scanPragmas collects per-file disabled rules via `#pragma lint:disable RULE[,RULE2]` lines.
// Enable pragmas (`#pragma lint:enable RULE`) remove disables within the same file.
func scanPragmas(dir, pkgRoot string) map[string]map[string]bool {
    disabled := map[string]map[string]bool{}
    root := filepath.Clean(filepath.Join(dir, pkgRoot))
    _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
        if err != nil || d.IsDir() || filepath.Ext(path) != ".ami" { return nil }
        f, err := os.Open(path)
        if err != nil { return nil }
        defer f.Close()
        rdr := bufio.NewScanner(f)
        for rdr.Scan() {
            line := strings.TrimSpace(rdr.Text())
            if !strings.HasPrefix(line, "#pragma lint:") { continue }
            line = strings.TrimPrefix(line, "#pragma lint:")
            if strings.HasPrefix(line, "disable ") {
                rules := strings.Split(strings.TrimSpace(strings.TrimPrefix(line, "disable ")), ",")
                for i := range rules { rules[i] = strings.TrimSpace(rules[i]) }
                for _, r := range rules {
                    if r == "" { continue }
                    m := disabled[path]
                    if m == nil { m = map[string]bool{}; disabled[path] = m }
                    m[r] = true
                }
            } else if strings.HasPrefix(line, "enable ") {
                rules := strings.Split(strings.TrimSpace(strings.TrimPrefix(line, "enable ")), ",")
                for i := range rules { rules[i] = strings.TrimSpace(rules[i]) }
                for _, r := range rules {
                    if r == "" { continue }
                    if m := disabled[path]; m != nil { delete(m, r) }
                }
            }
        }
        return nil
    })
    return disabled
}

