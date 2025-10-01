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
    mapIndexSet  = regexp.MustCompile(`^\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\[\s*\"([^\"]+)\"\s*\]\s*=`)
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
            // Track recent message variables assigned from fmt.Sprintf or literals
            recentMsgs := map[string]string{}
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
                    // Track message var := fmt.Sprintf("...") or := "literal"
                    if idx := strings.Index(line, ":="); idx > 0 {
                        name := strings.TrimSpace(line[:idx])
                        rhs := strings.TrimSpace(line[idx+2:])
                        if j := strings.Index(rhs, "fmt.Sprintf(\""); j >= 0 {
                            rest := rhs[j+len("fmt.Sprintf(\""):]
                            if k := strings.Index(rest, "\""); k >= 0 && name != "" { recentMsgs[name] = rest[:k] }
                        } else if strings.HasPrefix(rhs, "\"") {
                            lit := rhs[1:]
                            if k := strings.Index(lit, "\""); k >= 0 && name != "" { recentMsgs[name] = lit[:k] }
                        }
                    }
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
                    // Track bracket index assignments like data["paramName"] = ...
                    if m := mapIndexSet.FindStringSubmatch(line); m != nil && len(m) > 2 {
                        varName := m[1]
                        key := m[2]
                        for i := len(recentMaps) - 1; i >= 0; i-- {
                            if recentMaps[i].name == varName {
                                if recentMaps[i].keys == nil { recentMaps[i].keys = map[string]struct{}{} }
                                recentMaps[i].keys[key] = struct{}{}
                                break
                            }
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
                            } else {
                                // Variable form: Message: ident
                                // crude scan for 'Message:' and an identifier token
                                if i := strings.Index(content, "Message:"); i >= 0 {
                                    tail := strings.TrimSpace(content[i+len("Message:"):])
                                    if tail != "" {
                                        // cut at first comma or newline/brace
                                        cut := len(tail)
                                        if j := strings.IndexAny(tail, ",}\n"); j >= 0 { cut = j }
                                        ident := strings.TrimSpace(tail[:cut])
                                        // ensure it's an identifier (not quoted)
                                        if ident != "" && !strings.HasPrefix(ident, "\"") {
                                            if s, ok := recentMsgs[ident]; ok {
                                                if _, seen := samples[code]; !seen { samples[code] = s }
                                            }
                                        }
                                    }
                                }
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
    fmt.Fprintln(of, "- fieldPath: struct field traversal (e.g., a→b).")
    fmt.Fprintln(of, "- argIndex/tupleIndex: top-level call argument/return position. Summary 'paths' include these redundantly per entry.")
    fmt.Fprintln(of)
    fmt.Fprintln(of, "Examples:")
    fmt.Fprintln(of, "")
    fmt.Fprintln(of, "- E_GENERIC_ARITY_MISMATCH (return):")
    fmt.Fprintln(of, "  Code: E_GENERIC_ARITY_MISMATCH")
    fmt.Fprintln(of, "  Message: generic type argument count mismatch")
    fmt.Fprintln(of, "  Data:")
    fmt.Fprintln(of, "    function: \"F\"")
    fmt.Fprintln(of, "    index: 0")
    fmt.Fprintln(of, "    base: \"Owned\"")
    fmt.Fprintln(of, "    path: [\"slice\", \"Owned\"]")
    fmt.Fprintln(of, "    pathIdx: [0]")
    fmt.Fprintln(of, "    fieldPath: [\"a\", \"b\"]")
    fmt.Fprintln(of, "    expected: \"Struct{a:Struct{b:slice<Owned<T>>}}\"")
    fmt.Fprintln(of, "    actual: \"Struct{a:Struct{b:slice<Owned<int,string>>}}\"")
    fmt.Fprintln(of, "    expectedArity: 1")
    fmt.Fprintln(of, "    actualArity: 2")
    fmt.Fprintln(of, "    expectedPos: { line: 3, column: 16, offset: 42 }")
    fmt.Fprintln(of, "")
    fmt.Fprintln(of, "- E_RETURN_TUPLE_MISMATCH_SUMMARY:")
    fmt.Fprintln(of, "  Code: E_RETURN_TUPLE_MISMATCH_SUMMARY")
    fmt.Fprintln(of, "  Message: multiple return elements mismatch")
    fmt.Fprintln(of, "  Data:")
    fmt.Fprintln(of, "    function: \"F\"")
    fmt.Fprintln(of, "    count: 2")
    fmt.Fprintln(of, "    indices: [0, 1]")
    fmt.Fprintln(of, "    paths:")
    fmt.Fprintln(of, "      - index: 0, tupleIndex: 0, base: \"Owned\", path: [\"slice\", \"Owned\"], pathIdx: [0], fieldPath: [\"a\", \"b\"], expectedPos: { line: 3, column: 16, offset: 42 }")
    fmt.Fprintln(of, "      - index: 1, tupleIndex: 1, base: \"Error\", path: [\"Error\"], pathIdx: [], expectedPos: { line: 3, column: 27, offset: 64 }")
    fmt.Fprintln(of, "")
    fmt.Fprintln(of, "- E_CALL_ARGS_MISMATCH_SUMMARY:")
    fmt.Fprintln(of, "  Code: E_CALL_ARGS_MISMATCH_SUMMARY")
    fmt.Fprintln(of, "  Message: multiple call arguments mismatch")
    fmt.Fprintln(of, "  Data:")
    fmt.Fprintln(of, "    callee: \"Callee\"")
    fmt.Fprintln(of, "    count: 2")
    fmt.Fprintln(of, "    indices: [0, 1]")
    fmt.Fprintln(of, "    paths:")
    fmt.Fprintln(of, "      - argIndex: 0, paramName: \"a\", base: \"Owned\", path: [\"slice\", \"Owned\"], pathIdx: [0], fieldPath: [\"a\"], expectedPos: { line: 2, column: 16, offset: 26 }")
    fmt.Fprintln(of, "      - argIndex: 1, paramName: \"b\", base: \"Error\", path: [\"Error\"], pathIdx: [], expectedPos: { line: 2, column: 23, offset: 33 }")
}
