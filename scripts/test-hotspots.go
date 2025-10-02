//go:build ignore
// +build ignore

package main

import (
    "flag"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"
)

// This tool scans for testing hotspots:
//  - Enforced: packages (directories) under roots with .go files but zero *_test.go files.
//  - Advisory: individual .go files without a matching *_test.go sibling (same basename).
// Exit code is 1 if any package-level failures are detected; otherwise 0.
func main() {
    flag.Parse()
    roots := flag.Args()
    if len(roots) == 0 {
        roots = []string{"src"}
    }

    // Print banner for consistency with prior script
    fmt.Fprintln(os.Stderr, "Scanning src/ for test coverage hotspots...")

    // Exclude test fixture subtree (same as shell script)
    const excludePrefix = "src/cmd/ami/build/test/"

    type counts struct{ goFiles, testFiles int }
    byDir := map[string]*counts{}
    // Collect non-test .go files for advisory missing-pair checks
    var files []string

    for _, root := range roots {
        root = filepath.Clean(root)
        filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
            if err != nil { return err }
            sp := filepath.ToSlash(path)
            if d.IsDir() {
                if strings.HasPrefix(sp+"/", excludePrefix) {
                    return filepath.SkipDir
                }
                return nil
            }
            if !strings.HasSuffix(sp, ".go") { return nil }
            dir := filepath.Dir(sp)
            c := byDir[dir]
            if c == nil { c = &counts{}; byDir[dir] = c }
            c.goFiles++
            if strings.HasSuffix(sp, "_test.go") { c.testFiles++ } else {
                // Advisory list: ignore specific patterns like doc.go and *schema_embed.go
                base := filepath.Base(sp)
                if base == "doc.go" || strings.HasSuffix(sp, "schema_embed.go") {
                    return nil
                }
                if strings.HasPrefix(sp, excludePrefix) { return nil }
                files = append(files, sp)
            }
            return nil
        })
    }

    // Report packages with no *_test.go files
    var noTests int
    var missingPairs int
    dirs := make([]string, 0, len(byDir))
    for d := range byDir { dirs = append(dirs, d) }
    sort.Strings(dirs)
    for _, d := range dirs {
        c := byDir[d]
        if c.goFiles > 0 && c.testFiles == 0 {
            // Enforced failure
            fmt.Printf("NO_TESTS  %s\n", d)
            noTests++
        }
    }

    // Report files without a sibling *_test.go (advisory)
    sort.Strings(files)
    for _, f := range files {
        dir := filepath.ToSlash(filepath.Dir(f))
        base := strings.TrimSuffix(filepath.Base(f), ".go")
        expect := dir + "/" + base + "_test.go"
        if _, err := os.Stat(expect); err != nil {
            if os.IsNotExist(err) {
                fmt.Printf("MISSING_PAIR  %s  (expect: %s)\n", f, expect)
                missingPairs++
            } else {
                // If stat error other than not-exist, still surface advisory
                fmt.Printf("MISSING_PAIR  %s  (expect: %s; stat error: %v)\n", f, expect, err)
                missingPairs++
            }
        }
    }

    if noTests > 0 || missingPairs > 0 {
        if noTests > 0 {
            fmt.Fprintln(os.Stderr, "One or more packages have no tests.")
        }
        if missingPairs > 0 {
            fmt.Fprintf(os.Stderr, "One or more files are missing *_test.go pairs (count=%d).\n", missingPairs)
        }
        os.Exit(1)
    }
    os.Exit(0)
}
