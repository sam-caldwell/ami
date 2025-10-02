//go:build ignore
// +build ignore

package main

import (
    "bufio"
    "errors"
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
    modeFlag := flag.String("mode", "", "checking mode: cohesive|strict (default from CHECK_MODE or 'cohesive')")
    testModeFlag := flag.String("test", "", "test requirement: per-file|package (default from CHECK_TEST_MODE or 'per-file')")
    verboseFlag := flag.Bool("v", false, "verbose output")
    flag.Parse()

    mode := *modeFlag
    if mode == "" {
        mode = os.Getenv("CHECK_MODE")
    }
    if mode == "" { mode = "cohesive" }
    if mode != "cohesive" && mode != "strict" {
        fmt.Fprintf(os.Stderr, "ERROR: CHECK_MODE must be 'cohesive' or 'strict' (got %s)\n", mode)
        os.Exit(2)
    }

    testMode := *testModeFlag
    if testMode == "" { testMode = os.Getenv("CHECK_TEST_MODE") }
    if testMode == "" { testMode = "per-file" }
    if testMode != "per-file" && testMode != "package" {
        fmt.Fprintf(os.Stderr, "ERROR: CHECK_TEST_MODE must be 'per-file' or 'package' (got %s)\n", testMode)
        os.Exit(2)
    }

    verbose := *verboseFlag
    if !verbose {
        if os.Getenv("CHECK_VERBOSE") == "1" { verbose = true }
    }

    roots := flag.Args()
    if len(roots) == 0 { roots = []string{"."} }

    // Collect .go files, excluding vendor/build/.git
    var files []string
    for _, root := range roots {
        info, err := os.Stat(root)
        if err != nil {
            fmt.Fprintf(os.Stderr, "ERROR: path not found: %s\n", root)
            os.Exit(2)
        }
        if !info.IsDir() {
            if strings.HasSuffix(root, ".go") {
                files = append(files, root)
            }
            continue
        }
        filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
            if err != nil { return err }
            name := d.Name()
            if d.IsDir() {
                if name == ".git" || name == "vendor" || name == "build" { return filepath.SkipDir }
                return nil
            }
            if strings.HasSuffix(name, ".go") {
                files = append(files, path)
            }
            return nil
        })
    }
    sort.Strings(files)

    // Precompute per-directory test presence if package mode
    dirHasTests := map[string]bool{}
    if testMode == "package" {
        for _, p := range files {
            if strings.HasSuffix(p, "_test.go") {
                dirHasTests[filepath.Dir(p)] = true
            }
        }
    }

    fset := token.NewFileSet()
    var violations int
    for _, path := range files {
        res, err := analyzeFile(fset, path)
        if err != nil {
            fmt.Fprintf(os.Stderr, "%s: parse error: %v\n", path, err)
            violations++
            continue
        }
        // Skip docs/generated/embed-only
        if res.IsDoc || res.IsGenerated || res.HasGoEmbed {
            if verbose { fmt.Fprintf(os.Stderr, "skip: %s\n", path) }
            continue
        }
        // Single declaration check
        if !res.IsTest {
            if !checkSingleDecl(mode, res) {
                reportSingleDeclViolation(res)
                violations++
            }
            // Test presence check
            if requiresTest(res) {
                if testMode == "per-file" {
                    exp := strings.TrimSuffix(res.Path, ".go") + "_test.go"
                    if _, err := os.Stat(exp); err != nil {
                        if errors.Is(err, os.ErrNotExist) {
                            fmt.Printf("%s: missing corresponding test file %s\n", res.Path, exp)
                            violations++
                        } else {
                            fmt.Fprintf(os.Stderr, "%s: test check error: %v\n", res.Path, err)
                            violations++
                        }
                    }
                } else { // package
                    if !dirHasTests[filepath.Dir(res.Path)] {
                        fmt.Printf("%s: package has no *_test.go files\n", res.Path)
                        violations++
                    }
                }
            }
        }
    }
    if violations > 0 {
        fmt.Fprintf(os.Stderr, "Found %d violation(s)\n", violations)
        os.Exit(1)
    }
}

type Result struct {
    Path               string
    IsTest             bool
    IsDoc              bool
    IsGenerated        bool
    HasGoEmbed         bool
    TypeNames          []string
    MethodRecvNames    []string
    NonMethodFuncNames []string
    ConstTypedNames    []string // types referenced by typed const specs
    ConstBlocks        int
}

func analyzeFile(fset *token.FileSet, path string) (Result, error) {
    res := Result{Path: path, IsTest: strings.HasSuffix(path, "_test.go")}
    base := filepath.Base(path)
    if strings.EqualFold(base, "doc.go") { res.IsDoc = true }

    src, err := os.ReadFile(path)
    if err != nil { return res, err }
    content := string(src)
    if strings.Contains(content, "Code generated") && strings.Contains(content, "DO NOT EDIT") {
        res.IsGenerated = true
    }
    if strings.Contains(content, "//go:embed") {
        res.HasGoEmbed = true
    }

    f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
    if err != nil { return res, err }

    // Walk top-level declarations
    for _, d := range f.Decls {
        switch decl := d.(type) {
        case *ast.FuncDecl:
            if decl.Recv != nil && len(decl.Recv.List) > 0 {
                res.MethodRecvNames = append(res.MethodRecvNames, recvTypeName(decl.Recv.List[0].Type))
            } else {
                res.NonMethodFuncNames = append(res.NonMethodFuncNames, decl.Name.Name)
            }
        case *ast.GenDecl:
            switch decl.Tok {
            case token.TYPE:
                for _, s := range decl.Specs {
                    if ts, ok := s.(*ast.TypeSpec); ok {
                        res.TypeNames = append(res.TypeNames, ts.Name.Name)
                    }
                }
            case token.CONST:
                res.ConstBlocks++
                for _, s := range decl.Specs {
                    if vs, ok := s.(*ast.ValueSpec); ok && vs.Type != nil {
                        if tn := typeExprName(vs.Type); tn != "" { res.ConstTypedNames = append(res.ConstTypedNames, tn) }
                    }
                }
            default:
                // ignore imports/var
            }
        }
    }
    return res, nil
}

func checkSingleDecl(mode string, r Result) bool {
    switch mode {
    case "strict":
        // Exactly one of: any func (methods count individually), type spec, const block
        entities := len(r.TypeNames) + len(r.MethodRecvNames) + len(r.NonMethodFuncNames) + r.ConstBlocks
        return entities == 1
    default: // cohesive
        // Case 1: has exactly one type, allow methods for that type and consts typed as that type.
        if len(r.TypeNames) == 1 {
            t := r.TypeNames[0]
            // no non-method functions
            if len(r.NonMethodFuncNames) > 0 { return false }
            // all methods receivers match type
            for _, rt := range r.MethodRecvNames {
                if baseType(rt) != t { return false }
            }
            // if consts exist and are typed, they must match the type
            for _, ct := range r.ConstTypedNames {
                if baseType(ct) != t { return false }
            }
            return true
        }
        // Case 2: only methods, all receivers same type
        if len(r.TypeNames) == 0 && len(r.MethodRecvNames) > 0 && len(r.NonMethodFuncNames) == 0 {
            uniq := map[string]struct{}{}
            for _, rt := range r.MethodRecvNames { uniq[baseType(rt)] = struct{}{} }
            return len(uniq) == 1 && len(r.ConstTypedNames) == 0
        }
        // Case 3: single non-method function only
        if len(r.TypeNames) == 0 && len(r.MethodRecvNames) == 0 && len(r.NonMethodFuncNames) == 1 && r.ConstBlocks == 0 {
            return true
        }
        // Case 4: only const blocks (typed or untyped)
        if len(r.TypeNames) == 0 && len(r.MethodRecvNames) == 0 && len(r.NonMethodFuncNames) == 0 && r.ConstBlocks >= 1 {
            if len(r.ConstTypedNames) <= 1 { return true }
            uniq := map[string]struct{}{}
            for _, ct := range r.ConstTypedNames { uniq[baseType(ct)] = struct{}{} }
            return len(uniq) == 1
        }
        return false
    }
}

func requiresTest(r Result) bool {
    if r.IsDoc || r.IsGenerated || r.HasGoEmbed { return false }
    if r.IsTest { return false }
    if len(r.TypeNames) > 0 || len(r.MethodRecvNames) > 0 || len(r.NonMethodFuncNames) > 0 { return true }
    return false
}

func reportSingleDeclViolation(r Result) {
    var parts []string
    if n := len(r.TypeNames); n > 0 { parts = append(parts, fmt.Sprintf("types=%d[%s]", n, strings.Join(r.TypeNames, ","))) }
    if n := len(r.MethodRecvNames); n > 0 { parts = append(parts, fmt.Sprintf("methods=%d[%s]", n, strings.Join(r.MethodRecvNames, ","))) }
    if n := len(r.NonMethodFuncNames); n > 0 { parts = append(parts, fmt.Sprintf("funcs=%d[%s]", n, strings.Join(r.NonMethodFuncNames, ","))) }
    if r.ConstBlocks > 0 { parts = append(parts, fmt.Sprintf("const-blocks=%d", r.ConstBlocks)) }
    if len(parts) == 0 { parts = append(parts, "no-declarations") }
    fmt.Printf("%s: expected single cohesive declaration; found %s\n", r.Path, strings.Join(parts, ","))
}

func recvTypeName(expr ast.Expr) string {
    switch t := expr.(type) {
    case *ast.StarExpr:
        return recvTypeName(t.X)
    case *ast.Ident:
        return t.Name
    case *ast.IndexExpr:
        return recvTypeName(t.X)
    case *ast.IndexListExpr:
        return recvTypeName(t.X)
    case *ast.SelectorExpr:
        return t.Sel.Name
    default:
        return fmt.Sprintf("%T", expr)
    }
}

func typeExprName(expr ast.Expr) string {
    switch t := expr.(type) {
    case *ast.Ident:
        return t.Name
    case *ast.SelectorExpr:
        return t.Sel.Name
    case *ast.StarExpr:
        return typeExprName(t.X)
    case *ast.ArrayType:
        return typeExprName(t.Elt)
    default:
        return ""
    }
}

func baseType(s string) string { return strings.TrimLeft(s, "*[]") }

// readInputFiles exists to mirror earlier analyzer behavior if ever needed.
func readInputFiles() []string {
    var out []string
    s := bufio.NewScanner(os.Stdin)
    for s.Scan() { if v := strings.TrimSpace(s.Text()); v != "" { out = append(out, v) } }
    return out
}
