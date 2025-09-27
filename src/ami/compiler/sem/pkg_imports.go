package sem

import (
    "regexp"
    "time"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

var impPathRe = regexp.MustCompile(`^[A-Za-z0-9._\-/]+$`)

// AnalyzePackageAndImports emits diagnostics for invalid package identifiers and import paths.
// Rules (scaffold):
// - package name must start with a letter and contain only letters/digits (no underscores)
// - import paths must be non-empty and match a conservative character class
func AnalyzePackageAndImports(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // Package name checks
    if f.PackageName == "" || f.PackageName[0] < 'A' || (f.PackageName[0] > 'Z' && f.PackageName[0] < 'a') || f.PackageName[0] > 'z' {
        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PACKAGE_NAME_INVALID", Message: "invalid package name", Pos: &diag.Position{Line: f.PackagePos.Line, Column: f.PackagePos.Column, Offset: f.PackagePos.Offset}})
    }
    for i := 0; i < len(f.PackageName); i++ {
        ch := f.PackageName[i]
        if !(ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9') {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PACKAGE_NAME_INVALID", Message: "invalid package name characters", Pos: &diag.Position{Line: f.PackagePos.Line, Column: f.PackagePos.Column, Offset: f.PackagePos.Offset}})
            break
        }
    }
    // Import paths
    for _, d := range f.Decls {
        if im, ok := d.(*ast.ImportDecl); ok {
            if im.Path == "" || !impPathRe.MatchString(im.Path) {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_IMPORT_PATH_INVALID", Message: "invalid import path", Pos: &diag.Position{Line: im.PathPos.Line, Column: im.PathPos.Column, Offset: im.PathPos.Offset}})
                continue
            }
            // stricter rules: not absolute, no traversal, no empty segments, no trailing slash
            p := im.Path
            if len(p) > 0 && p[0] == '/' {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_IMPORT_PATH_ABSOLUTE", Message: "import path must not be absolute", Pos: &diag.Position{Line: im.PathPos.Line, Column: im.PathPos.Column, Offset: im.PathPos.Offset}})
            }
            if strings.Contains(p, "//") {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_IMPORT_PATH_EMPTY_SEGMENT", Message: "import path contains empty segment", Pos: &diag.Position{Line: im.PathPos.Line, Column: im.PathPos.Column, Offset: im.PathPos.Offset}})
            }
            if strings.HasSuffix(p, "/") {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_IMPORT_PATH_TRAILING_SLASH", Message: "import path must not end with '/'", Pos: &diag.Position{Line: im.PathPos.Line, Column: im.PathPos.Column, Offset: im.PathPos.Offset}})
            }
            if strings.Contains(p, "/../") || strings.HasPrefix(p, "../") || strings.HasSuffix(p, "/..") {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_IMPORT_PATH_TRAVERSAL", Message: "import path must not traverse", Pos: &diag.Position{Line: im.PathPos.Line, Column: im.PathPos.Column, Offset: im.PathPos.Offset}})
            }
        }
    }
    return out
}
