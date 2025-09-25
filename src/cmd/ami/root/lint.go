package root

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"
    "regexp"

    "github.com/spf13/cobra"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/sem"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/internal/logger"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

func newLintCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "lint",
        Short: "Lint AMI sources in the workspace",
        Example: `  ami lint
  ami --json lint`,
        Run: func(cmd *cobra.Command, args []string) {
            runLint()
        },
    }
}

// runLint discovers AMI source entrypoints using the workspace, orders
// them so that imported workspace packages are linted before the main
// package, and emits a sources.v1 summary (JSON when --json).
func runLint() {
    // Load workspace; on failure, emit diagnostic consistent with build.
    const wsPath = "ami.workspace"
    ws, err := workspace.Load(wsPath)
    if err != nil {
        if flagJSON {
            d := sch.DiagV1{Schema: "diag.v1", Timestamp: sch.FormatTimestamp(nowUTC()), Level: "error", Code: "E_WS_SCHEMA", Message: fmt.Sprintf("workspace validation failed: %v", err), File: wsPath}
            if verr := d.Validate(); verr == nil {
                _ = json.NewEncoder(os.Stdout).Encode(d)
            }
        } else {
            fmt.Fprintln(os.Stderr, fmt.Sprintf("workspace error: %v", err))
        }
        return
    }

    // Derive package -> root map from workspace
    pkgRoots := parseWorkspacePackages(ws)
    // Workspace package rules (names, versions)
    diags := lintWorkspacePackages(ws)
    // Discover entry units ('.ami' files) per package
    // For main, prefer src/main.ami when present.
    // Lint order: imports (workspace-local) first, then main.
    order := lintOrder(pkgRoots)

    // Build sources summary in the discovered order
    sources := sch.SourcesV1{Schema: "sources.v1", Timestamp: sch.FormatTimestamp(nowUTC())}
    // Deterministic file collection per package (sorted)
    type unit struct{ pkg, file, src string; imports []string; ast *astpkg.File }
    var ulist []unit
    for _, pkg := range order {
        root := pkgRoots[pkg]
        files, _ := filepath.Glob(filepath.Join(root, "*.ami"))
        sort.Strings(files)
        for _, f := range files {
            b, _ := os.ReadFile(f)
            src := string(b)
            imports := parser.ExtractImports(src)
            // keep AST for lint rules
            p := parser.New(src)
            ast := p.ParseFile()
            ulist = append(ulist, unit{pkg: pkg, file: f, src: src, imports: imports, ast: ast})
            sources.Units = append(sources.Units, sch.SourceUnit{Package: pkg, File: f, Imports: imports, Source: src})
        }
    }

    // Lint diagnostics across units
    for _, u := range ulist {
        diags = append(diags, lintUnit(u.pkg, u.file, u.src, u.ast)...)
    }

    // Output
    if flagJSON {
        // Emit sources summary first for deterministic tooling
        if err := sources.Validate(); err == nil { _ = json.NewEncoder(os.Stdout).Encode(sources) }
        // Emit lint diagnostics as JSON lines
        errs := 0; warns := 0
        for _, d := range diags {
            if d.Level == diag.Error { errs++ } else if d.Level == diag.Warn { warns++ }
            sd := d.ToSchema()
            if sd.Package == "" { sd.Package = uPackageFromPath(d.File, pkgRoots) }
            _ = json.NewEncoder(os.Stdout).Encode(sd)
        }
        // Summary record
        summary := sch.DiagV1{Schema: "diag.v1", Timestamp: sch.FormatTimestamp(nowUTC()), Level: "info", Code: "LINT_SUMMARY", Message: fmt.Sprintf("%d warnings, %d errors", warns, errs)}
        _ = json.NewEncoder(os.Stdout).Encode(summary)
        return
    }
    // Human mode: concise, ordered list with filenames inline
    logger.Info(fmt.Sprintf("lint: discovered %d units", len(sources.Units)), nil)
    for _, u := range sources.Units {
        logger.Info(fmt.Sprintf("lint: unit %s (%s)", u.File, u.Package), nil)
    }
    // Human diagnostics
    errs := 0; warns := 0
    for _, d := range diags {
        msg := fmt.Sprintf("%s: %s", d.Code, d.Message)
        if d.File != "" { msg = d.File + ": " + msg }
        switch d.Level {
        case diag.Error:
            logger.Error(msg, nil); errs++
        case diag.Warn:
            logger.Warn(msg, nil); warns++
        default:
            logger.Info(msg, nil)
        }
    }
    logger.Info(fmt.Sprintf("lint: summary: %d warnings, %d errors", warns, errs), nil)
}

// parseWorkspacePackages extracts a map of package name -> root directory
// from the loosely-typed workspace.Packages field.
func parseWorkspacePackages(ws *workspace.Workspace) map[string]string {
    out := map[string]string{}
    for _, p := range ws.Packages {
        m, ok := p.(map[string]any)
        if !ok { continue }
        for name, v := range m {
            // value is expected to be a map with at least 'root'
            if vm, ok := v.(map[string]any); ok {
                if r, ok := vm["root"].(string); ok && strings.TrimSpace(r) != "" {
                    out[name] = r
                }
            }
        }
    }
    // If no packages parsed, default to main: ./src (common scaffold)
    if len(out) == 0 {
        out["main"] = "./src"
    }
    return out
}

// lintOrder returns packages in the order they should be linted:
// imported workspace-local packages first, then main. If a local
// package itself imports other local packages, those are ordered
// before it (DFS). Packages without resolvable imports are included
// once. Unknown imports (external) are ignored for ordering.
func lintOrder(pkgRoots map[string]string) []string {
    // Helper to read a representative unit for a package to extract imports.
    importsFor := func(pkg string) []string {
        root, ok := pkgRoots[pkg]
        if !ok { return nil }
        // Prefer main.ami; otherwise first *.ami by name.
        mainPath := filepath.Join(root, "main.ami")
        var path string
        if fi, err := os.Stat(mainPath); err == nil && !fi.IsDir() {
            path = mainPath
        } else {
            list, _ := filepath.Glob(filepath.Join(root, "*.ami"))
            sort.Strings(list)
            if len(list) > 0 { path = list[0] }
        }
        if path == "" { return nil }
        b, err := os.ReadFile(path)
        if err != nil { return nil }
        return parser.ExtractImports(string(b))
    }

    // DFS
    visited := map[string]bool{}
    order := []string{}
    var visit func(string)
    visit = func(pkg string) {
        if visited[pkg] { return }
        visited[pkg] = true
        // For each import that is a workspace-local package, visit first
        for _, imp := range importsFor(pkg) {
            if _, ok := pkgRoots[imp]; ok {
                visit(imp)
            }
        }
        order = append(order, pkg)
    }
    // Always start from main if present
    if _, ok := pkgRoots["main"]; ok {
        visit("main")
    }
    // Include any remaining packages that weren't reachable from main
    // in stable order (sorted by name)
    var rest []string
    for k := range pkgRoots {
        if !visited[k] { rest = append(rest, k) }
    }
    sort.Strings(rest)
    for _, k := range rest { visit(k) }
    return order
}

// lintWorkspacePackages enforces workspace-level package naming and version rules.
func lintWorkspacePackages(ws *workspace.Workspace) []diag.Diagnostic {
    var out []diag.Diagnostic
    semverRe := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)
    for _, p := range ws.Packages {
        m, ok := p.(map[string]any)
        if !ok { continue }
        for name, v := range m {
            // package key name should be a valid import path (Chapter 3.0 composition)
            if !parser.ValidateImportPath(name) {
                out = append(out, diag.Diagnostic{Level: diag.Error, Code: "E_WS_PKG_NAME", Message: fmt.Sprintf("invalid workspace package name: %q", name)})
            }
            vm, ok := v.(map[string]any)
            if !ok { continue }
            if ver, ok := vm["version"].(string); ok {
                if !semverRe.MatchString(ver) {
                    out = append(out, diag.Diagnostic{Level: diag.Error, Code: "E_WS_PKG_VERSION", Message: fmt.Sprintf("package %q has invalid semantic version: %q", name, ver)})
                }
            }
        }
    }
    return out
}

// lintUnit returns diagnostics for a single unit.
func lintUnit(pkgName, filePath, src string, f *astpkg.File) []diag.Diagnostic {
    var out []diag.Diagnostic
    // Parser-level errors captured already during AST creation
    p := parser.New(src)
    _ = p.ParseFile()
    for _, e := range p.Errors() {
        if e.File == "" || e.File == "input.ami" { e.File = filePath }
        if e.Package == "" { e.Package = pkgName }
        out = append(out, e)
    }
    // Semantic analyzer
    semres := sem.AnalyzeFile(f)
    for _, e := range semres.Diagnostics {
        if e.File == "" { e.File = filePath }
        if e.Package == "" { e.Package = pkgName }
        out = append(out, e)
    }
    // Naming: package should be lowercase per style
    if f.Package != strings.ToLower(f.Package) && f.Package != "" {
        out = append(out, diag.Diagnostic{Level: diag.Warn, Code: "W_PKG_LOWERCASE", Message: "package name should be lowercase", Package: pkgName, File: filePath})
    }
    // Imports: duplicates and unused
    // Collect imports with alias
    type impInfo struct{ path, alias string }
    var imports []impInfo
    seenPath := map[string]bool{}
    for _, d := range f.Decls {
        if id, ok := d.(astpkg.ImportDecl); ok {
            alias := id.Alias
            if alias == "" {
                alias = pathBase(id.Path)
            }
            imports = append(imports, impInfo{path: id.Path, alias: alias})
            if seenPath[id.Path] {
                out = append(out, diag.Diagnostic{Level: diag.Warn, Code: "W_DUP_IMPORT", Message: fmt.Sprintf("duplicate import %q", id.Path), Package: pkgName, File: filePath})
            }
            seenPath[id.Path] = true
        }
    }
    // Build used identifier set from function bodies and pipeline args
    used := map[string]bool{}
    for _, d := range f.Decls {
        if fd, ok := d.(astpkg.FuncDecl); ok {
            for _, t := range fd.Body {
                if t.Kind == tok.IDENT { used[t.Lexeme] = true }
            }
        }
        if pd, ok := d.(astpkg.PipelineDecl); ok {
            markArgs := func(steps []astpkg.NodeCall) {
                for _, st := range steps {
                    for _, a := range st.Args {
                        for _, tok := range splitIdents(a) { used[tok] = true }
                    }
                }
            }
            markArgs(pd.Steps)
            markArgs(pd.ErrorSteps)
        }
    }
    for _, im := range imports {
        if im.alias == "_" { continue } // parser/sem already error; skip here
        if !used[im.alias] {
            out = append(out, diag.Diagnostic{Level: diag.Warn, Code: "W_UNUSED_IMPORT", Message: fmt.Sprintf("import %q (%s) not used", im.path, im.alias), Package: pkgName, File: filePath})
        }
    }
    // Formatting: final newline and CRLF
    if !strings.HasSuffix(src, "\n") {
        out = append(out, diag.Diagnostic{Level: diag.Warn, Code: "W_FILE_NO_NEWLINE", Message: "file does not end with newline", Package: pkgName, File: filePath})
    }
    if strings.Contains(src, "\r\n") {
        out = append(out, diag.Diagnostic{Level: diag.Warn, Code: "W_FILE_CRLF", Message: "file contains CRLF line endings; use LF", Package: pkgName, File: filePath})
    }
    return out
}

func pathBase(p string) string {
    if i := strings.LastIndex(p, "/"); i >= 0 { return p[i+1:] }
    return p
}

// splitIdents returns identifier-like tokens from a string (A-Za-z0-9_)
func splitIdents(s string) []string {
    var out []string
    cur := strings.Builder{}
    for _, r := range s {
        if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
            cur.WriteRune(r)
        } else {
            if cur.Len() > 0 { out = append(out, cur.String()); cur.Reset() }
        }
    }
    if cur.Len() > 0 { out = append(out, cur.String()) }
    return out
}

// uPackageFromPath maps a file path back to a package name using known roots.
func uPackageFromPath(file string, roots map[string]string) string {
    // Attempt: if file path has a root prefix, return that package
    // Iterate stable by package name for determinism
    var names []string
    for name := range roots { names = append(names, name) }
    sort.Strings(names)
    for _, name := range names {
        root := filepath.Clean(roots[name])
        if strings.HasPrefix(filepath.Clean(file), filepath.Clean(root)+string(os.PathSeparator)) {
            return name
        }
    }
    return ""
}
