package main

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "sort"
    "strings"
)

// A lightweight generator that scans Go source for diag.Record literals and extracts
// Code and Data keys to produce docs/diag-codes.md. Not a full Go AST parser; relies
// on simple regex patterns and is best-effort.

var (
    codeRe       = regexp.MustCompile(`Code:\s*\"([A-Z]_[A-Z0-9_]+)\"`)
    dataKeyRe    = regexp.MustCompile(`Data:\s*map\[string\]any\{([^}]*)\}`)
    kvRe         = regexp.MustCompile(`\"([a-zA-Z0-9_]+)\"\s*:`)
    msgSimpleRe  = regexp.MustCompile(`Message:\s*\"([^\"]+)\"`)
    msgFmtRe     = regexp.MustCompile(`Message:\s*fmt\.Sprintf\(\"([^\"]+)\"`)
    dataVarRe    = regexp.MustCompile(`Data:\s*([a-zA-Z_][a-zA-Z0-9_]*)`)
)

func main() {
    repoRoot := "."
    codes := map[string]map[string]struct{}{}
    samples := map[string]string{}
    // Scan selected dirs
    roots := []string{"src/ami/compiler/sem", "src/cmd/ami"}
    for _, root := range roots {
        _ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
            if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") { return nil }
            f, err := os.Open(path); if err != nil { return nil }
            defer f.Close()
            scanner := bufio.NewScanner(f)
            inRec := false
            braceDepth := 0
            var recBuf strings.Builder
            // Track most recent map[string]any assignments to variables (e.g., `data := map[string]any{...}`)
            type mapAssign struct{ name string; keys map[string]struct{} }
            var currentMapVar string
            mapBraceDepth := 0
            inMapAssign := false
            recentMaps := []mapAssign{}
            for scanner.Scan() {
                line := scanner.Text()
                // track record start/brace depth
                if strings.Contains(line, "diag.Record{") {
                    inRec = true
                    recBuf.Reset()
                    recBuf.WriteString(line)
                    recBuf.WriteByte('\n')
                    braceDepth = strings.Count(line, "{") - strings.Count(line, "}")
                    continue
                }
                // Track variable map assignments like `data := map[string]any{ ... }`
                if !inRec {
                    // Start of assignment
                    if !inMapAssign && strings.Contains(line, ":= map[string]any{") {
                        // Extract var name before ':='
                        idx := strings.Index(line, ":= map[string]any{")
                        if idx > 0 {
                            name := strings.TrimSpace(line[:idx])
                            currentMapVar = name
                            inMapAssign = true
                            mapBraceDepth = strings.Count(line, "{") - strings.Count(line, "}")
                            // seed recent map with any keys on same line after '{'
                            keys := map[string]struct{}{}
                            if cm := dataKeyRe.FindStringSubmatch("Data: " + line[idx:]); cm != nil && len(cm) > 1 {
                                for _, km := range kvRe.FindAllStringSubmatch(cm[1], -1) {
                                    if len(km) > 1 { keys[km[1]] = struct{}{} }
                                }
                            }
                            recentMaps = append(recentMaps, mapAssign{name: currentMapVar, keys: keys})
                            continue
                        }
                    }
                    if inMapAssign {
                        mapBraceDepth += strings.Count(line, "{") - strings.Count(line, "}")
                        if len(recentMaps) > 0 {
                            // accumulate keys
                            for _, km := range kvRe.FindAllStringSubmatch(line, -1) {
                                if len(km) > 1 { recentMaps[len(recentMaps)-1].keys[km[1]] = struct{}{} }
                            }
                        }
                        if mapBraceDepth <= 0 {
                            inMapAssign = false
                            currentMapVar = ""
                        }
                    }
                }
                if inRec {
                    braceDepth += strings.Count(line, "{") - strings.Count(line, "}")
                    recBuf.WriteString(line)
                    recBuf.WriteByte('\n')
                    if braceDepth <= 0 {
                        // parse this record buffer
                        content := recBuf.String()
                        // code
                        code := ""
                        if cm := codeRe.FindStringSubmatch(content); cm != nil { code = cm[1] }
                        if code != "" {
                            if _, ok := codes[code]; !ok { codes[code] = map[string]struct{}{} }
                            // data keys (may be multiple Data maps; collect all keys)
                            for _, dm := range dataKeyRe.FindAllStringSubmatch(content, -1) {
                                if len(dm) > 1 {
                                    for _, km := range kvRe.FindAllStringSubmatch(dm[1], -1) {
                                        if len(km) > 1 { codes[code][km[1]] = struct{}{} }
                                    }
                                }
                            }
                            // If Data references a var (e.g., Data: data), try to source keys from recent map assignment
                            if m := dataVarRe.FindStringSubmatch(content); m != nil && len(m) > 1 {
                                varName := m[1]
                                // Find the most recent assignment with this var name (scan from end)
                                for i := len(recentMaps) - 1; i >= 0; i-- {
                                    if recentMaps[i].name == varName {
                                        for k := range recentMaps[i].keys { codes[code][k] = struct{}{} }
                                        break
                                    }
                                }
                            }
                            // message sample: prefer fmt.Sprintf format; fallback to simple literal
                            if sm := msgFmtRe.FindStringSubmatch(content); sm != nil {
                                if _, seen := samples[code]; !seen { samples[code] = sm[1] }
                            } else if sm := msgSimpleRe.FindStringSubmatch(content); sm != nil {
                                if _, seen := samples[code]; !seen { samples[code] = sm[1] }
                            }
                        }
                        // reset
                        inRec = false
                        recBuf.Reset()
                    }
                }
            }
            return nil
        })
    }
    // Render
    outPath := filepath.Join(repoRoot, "docs", "diag-codes.md")
    _ = os.MkdirAll(filepath.Dir(outPath), 0o755)
    of, err := os.Create(outPath)
    if err != nil { fmt.Fprintln(os.Stderr, "create diag-codes.md:", err); os.Exit(1) }
    defer of.Close()
    fmt.Fprintln(of, "## Diagnostic Codes and Data Keys (Generated)")
    fmt.Fprintln(of, "")
    fmt.Fprintln(of, "Note: generated by tools/gen-diag-codes; best-effort from code literals.")
    fmt.Fprintln(of, "")
    // Sort codes for determinism
    var codeList []string
    for c := range codes { codeList = append(codeList, c) }
    sort.Strings(codeList)
    for _, c := range codeList {
        fmt.Fprintf(of, "- %s", c)
        // message sample
        if msg, ok := samples[c]; ok && strings.TrimSpace(msg) != "" {
            fmt.Fprintf(of, ": message sample = %q", msg)
        }
        keys := make([]string, 0, len(codes[c]))
        for k := range codes[c] { keys = append(keys, k) }
        sort.Strings(keys)
        if len(keys) > 0 {
            if _, ok := samples[c]; ok { fmt.Fprintf(of, "; ") } else { fmt.Fprintf(of, ": ") }
            fmt.Fprintf(of, "data keys = %s", strings.Join(keys, ", "))
        }
        fmt.Fprintln(of)
    }
    // Append brief notes for readers
    fmt.Fprintln(of)
    fmt.Fprintln(of, "Notes:")
    fmt.Fprintln(of, "- message sample: example literal or fmt.Sprintf format; dynamic parts omitted.")
    fmt.Fprintln(of, "- data keys: union of keys seen across code occurrences.")
    fmt.Fprintln(of, "- path/pathIdx: generic base nesting and argument indices, outer→inner.")
    fmt.Fprintln(of, "- fieldPath: struct field traversal (e.g., Struct→a→Struct→b).")
    fmt.Fprintln(of, "- argIndex/tupleIndex: top-level call argument/return position. Summary 'paths' include these redundantly per entry.")
}
