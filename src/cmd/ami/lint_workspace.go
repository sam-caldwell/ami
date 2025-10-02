package main

import (
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/semver"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
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
                } else if constraint != "" {
                    // Local import with version constraint: ensure the local package version satisfies the constraint
                    if c, err := semver.ParseConstraint(constraint); err == nil {
                        if p := findPackageByRoot(ws, path); p != nil {
                            if !semver.Satisfies(p.Version, c) {
                                diags = append(diags, diag.Record{Timestamp: now, Level: diag.Warn, Code: "E_IMPORT_CONSTRAINT", Message: "local package version does not satisfy import constraint: " + entry, File: "ami.workspace"})
                            }
                        }
                    }
                }
            }
        }
    }
    // Order check: lexical order by normalized path
    if !isSorted(normalizeForOrder(norm)) {
        diags = append(diags, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_IMPORT_ORDER", Message: "imports not sorted", File: "ami.workspace"})
    }

    // Validate package versions are valid SemVer (workspace declarations)
    for _, pe := range ws.Packages {
        v := strings.TrimSpace(pe.Package.Version)
        if v == "" || !semver.ValidateVersion(v) {
            diags = append(diags, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_PKG_VERSION_INVALID", Message: "package version is not valid semver: " + v, File: "ami.workspace"})
        }
    }
    return diags
}

