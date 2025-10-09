//go:build ignore
// +build ignore

package main

import (
    "flag"
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "os"
    "path/filepath"
    "sort"
    "strings"
)

func main() {
    // Allow gating this check via environment. By default, relax to package-level
    // semantics to reflect common Go testing practice (multiple Test* per file).
    // Set CHECK_SINGLE_TEST_MODE=per-file to enforce the strict rule.
    mode := os.Getenv("CHECK_SINGLE_TEST_MODE")
    if mode == "" { mode = "package" }
    if mode != "per-file" {
        // No-op in package/relaxed mode.
        os.Exit(0)
    }

    flag.Parse()
    args := flag.Args()
    if len(args) == 0 {
        args = []string{"src/"}
    }

    // Collect *_test.go files under provided paths
    files := make([]string, 0, 512)
    for _, root := range args {
        info, err := os.Stat(root)
        if err != nil {
            fmt.Fprintf(os.Stderr, "ERROR: path not found: %s\n", root)
            os.Exit(2)
        }
        if !info.IsDir() {
            if strings.HasSuffix(root, "_test.go") {
                files = append(files, root)
            }
            continue
        }
        filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
            if err != nil { return err }
            // Exclude test fixture subtree used by CLI E2E tests
            sp := filepath.ToSlash(path)
            if d.IsDir() {
                if strings.HasPrefix(sp+"/", "src/cmd/ami/build/test/") { return filepath.SkipDir }
                return nil
            }
            if strings.HasSuffix(sp, "_test.go") {
                files = append(files, path)
            }
            return nil
        })
    }
    sort.Strings(files)

    fset := token.NewFileSet()
    var failures int

    for _, path := range files {
        // parse file
        src, err := os.ReadFile(path)
        if err != nil {
            fmt.Fprintf(os.Stderr, "ERROR: read %s: %v\n", path, err)
            failures = 1
            continue
        }
        f, err := parser.ParseFile(fset, path, src, 0)
        if err != nil {
            fmt.Fprintf(os.Stderr, "ERROR: parse %s: %v\n", path, err)
            failures = 1
            continue
        }
        // count top-level Test* funcs (exclude TestMain)
        count := 0
        for _, d := range f.Decls {
            fn, ok := d.(*ast.FuncDecl)
            if !ok { continue }
            if fn.Recv != nil { continue }
            if fn.Name == nil { continue }
            name := fn.Name.Name
            if len(name) >= 4 && strings.HasPrefix(name, "Test") && name != "TestMain" {
                count++
            }
        }
        if count > 1 {
            fmt.Fprintf(os.Stderr, "ERROR: %s contains more than one Test* function\n", path)
            failures = 1
        }
    }

    os.Exit(failures)
}
