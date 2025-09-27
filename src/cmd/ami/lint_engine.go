package main

import (
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "time"

    diag "github.com/sam-caldwell/ami/src/schemas/diag"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    "github.com/sam-caldwell/ami/src/ami/semver"
)

var pkgNameRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`)
var importLexemeRe = regexp.MustCompile(`^[A-Za-z0-9._/@:+-]+$`)
var semverRe = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$`)
var constraintRe = regexp.MustCompile(`^(?:\^|~|>=|>)?\s*(?:v?\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?)$|^==latest$`)

// lintWorkspace runs simple naming/import checks based on the workspace manifest.
func lintWorkspace(dir string, ws *workspace.Workspace) []diag.Record {
    var diags []diag.Record
    now := time.Now().UTC()

    // Require main package and valid name.
    mainPkg := ws.FindPackage("main")
    if mainPkg == nil {
        diags = append(diags, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_WS_SCHEMA", Message: "missing main package", File: "ami.workspace"})
        return diags
    }
    if !validPkgName(mainPkg.Name) {
        diags = append(diags, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PKG_NAME_STYLE", Message: "package name should be PascalCase, camelCase, or lowercase without underscores", File: "ami.workspace"})
    }

    // Validate import lexemes and local paths.
    seen := map[string]bool{}
    // Track order
    norm := make([]string, 0, len(mainPkg.Import))
    for _, entry := range mainPkg.Import {
        path, constraint := splitImportConstraint(entry)
        if path == "" || !importLexemeRe.MatchString(path) {
            diags = append(diags, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_IMPORT_SYNTAX", Message: "import entry has invalid characters", File: "ami.workspace"})
            continue
        }
        if seen[path] {
            diags = append(diags, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_IMPORT_DUPLICATE", Message: "duplicate import entry: " + entry, File: "ami.workspace"})
        }
        seen[path] = true
        norm = append(norm, path)
        // Validate constraint if present
        if constraint != "" && !semver.ValidateConstraint(constraint) {
            diags = append(diags, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_IMPORT_CONSTRAINT_INVALID", Message: "invalid version constraint: " + constraint, File: "ami.workspace"})
        }
        // Local path check: starts with ./ or ../
        if stringsHasPrefixAny(path, []string{"./", "../"}) {
            if stringsHasPrefixAny(entry, []string{"../"}) {
                diags = append(diags, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_IMPORT_RELATIVE", Message: "relative parent path not allowed: " + entry, File: "ami.workspace"})
                continue
            }
            // Ensure exists and within workspace dir
            abs := filepath.Clean(filepath.Join(dir, path))
            if _, err := os.Stat(abs); err != nil {
                diags = append(diags, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_IMPORT_LOCAL_MISSING", Message: "local import path not found: " + entry, File: "ami.workspace"})
            } else {
                // Path exists; ensure it is declared as a package in workspace
                if findPackageByRoot(ws, path) == nil {
                    diags = append(diags, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_IMPORT_LOCAL_UNDECLARED", Message: "local import not declared as package: " + entry, File: "ami.workspace"})
                }
            }
        }
    }
    // Order check: lexical order by normalized path
    if !isSorted(normalizeForOrder(norm)) {
        diags = append(diags, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_IMPORT_ORDER", Message: "imports not sorted", File: "ami.workspace"})
    }
    return diags
}

// findPackageByRoot returns the workspace package whose Root matches the given path.
// The comparison is done on the raw Root string as declared in the workspace (paths
// are expected to be workspace-relative like ./lib). Returns nil when not found.
func findPackageByRoot(ws *workspace.Workspace, root string) *workspace.Package {
    for i := range ws.Packages {
        if ws.Packages[i].Package.Root == root {
            return &ws.Packages[i].Package
        }
    }
    return nil
}

// collectLocalImportRoots walks local (./...) imports recursively starting from pkg,
// returning a DFS order with children before parents (child-first) and duplicates eliminated.
func collectLocalImportRoots(ws *workspace.Workspace, pkg *workspace.Package) []string {
    visited := make(map[string]bool)
    var order []string
    var dfs func(p *workspace.Package)
    dfs = func(p *workspace.Package) {
        // Traverse each local import
        for _, entry := range p.Import {
            path, _ := splitImportConstraint(entry)
            if !stringsHasPrefixAny(path, []string{"./"}) || stringsHasPrefixAny(path, []string{"../"}) {
                continue
            }
            if visited[path] {
                continue
            }
            visited[path] = true
            if child := findPackageByRoot(ws, path); child != nil {
                dfs(child)
            }
            order = append(order, path)
        }
    }
    dfs(pkg)
    return order
}

// stringsHasPrefixAny reports whether s has any of the given prefixes.
func stringsHasPrefixAny(s string, prefixes []string) bool {
    for _, p := range prefixes {
        if len(s) >= len(p) && s[:len(p)] == p {
            return true
        }
    }
    return false
}

// validPkgName returns true when name conforms to allowed styles: lowercase, camelCase, or PascalCase; underscores disallowed.
func validPkgName(name string) bool {
    if name == "" { return false }
    if containsUnderscore(name) { return false }
    // lowercase
    if regexp.MustCompile(`^[a-z][a-z0-9]*$`).MatchString(name) { return true }
    // camelCase
    if regexp.MustCompile(`^[a-z][A-Za-z0-9]*$`).MatchString(name) { return true }
    // PascalCase
    if regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`).MatchString(name) { return true }
    return false
}

func containsUnderscore(s string) bool {
    for i := 0; i < len(s); i++ { if s[i] == '_' { return true } }
    return false
}

// splitImportConstraint splits an import entry like "path >= v1.2.3" into
// path and constraint parts. When no constraint is recognized, returns the
// original entry and empty constraint.
func splitImportConstraint(entry string) (string, string) {
    // Try to split by whitespace; treat the first token as path and the remainder as constraint.
    parts := strings.Fields(entry)
    if len(parts) <= 1 {
        return entry, ""
    }
    path := parts[0]
    constraint := strings.TrimSpace(strings.Join(parts[1:], " "))
    return path, constraint
}

func isSorted(ss []string) bool {
    for i := 1; i < len(ss); i++ { if ss[i-1] > ss[i] { return false } }
    return true
}

func normalizeForOrder(ss []string) []string {
    out := make([]string, len(ss))
    for i, s := range ss {
        s = strings.TrimPrefix(s, "./")
        s = strings.TrimSuffix(s, "/")
        out[i] = strings.ToLower(s)
    }
    return out
}
