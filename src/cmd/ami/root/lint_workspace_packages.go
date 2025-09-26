package root

import (
    "fmt"
    "regexp"

    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// lintWorkspacePackages enforces workspace-level package naming and version rules.
func lintWorkspacePackages(ws *workspace.Workspace) []diag.Diagnostic {
    var out []diag.Diagnostic
    semverRe := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)
    for _, p := range ws.Packages {
        m, ok := p.(map[string]any)
        if !ok {
            continue
        }
        for name, v := range m {
            // package key name should be a valid import path
            if !parser.ValidateImportPath(name) {
                out = append(out, diag.Diagnostic{Level: diag.Error, Code: "E_WS_PKG_NAME", Message: fmt.Sprintf("invalid workspace package name: %q", name)})
            }
            vm, ok := v.(map[string]any)
            if !ok {
                continue
            }
            if ver, ok := vm["version"].(string); ok {
                if !semverRe.MatchString(ver) {
                    out = append(out, diag.Diagnostic{Level: diag.Error, Code: "E_WS_PKG_VERSION", Message: fmt.Sprintf("package %q has invalid semantic version: %q", name, ver)})
                }
            }
        }
    }
    return out
}
