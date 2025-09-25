package mod

import (
    "errors"
    "os"
    "path/filepath"
    "strings"
)

// findWorkspaceRoot walks up from start looking for ami.workspace.
func findWorkspaceRoot(start string) (string, error) {
    dir := start
    for {
        if _, err := os.Stat(filepath.Join(dir, "ami.workspace")); err == nil {
            return dir, nil
        }
        parent := filepath.Dir(dir)
        if parent == dir { // reached filesystem root
            return "", errors.New("ami.workspace not found in current or parent directories")
        }
        dir = parent
    }
}

// parseWorkspaceImports scans ami.workspace for import items using a lightweight parser.
func parseWorkspaceImports(wsPath string) ([]string, error) {
    b, err := os.ReadFile(wsPath)
    if err != nil { return nil, err }
    lines := strings.Split(string(b), "\n")
    inImport := false
    imports := []string{}
    for _, ln := range lines {
        s := strings.TrimSpace(ln)
        if strings.HasPrefix(s, "import:") { inImport = true; continue }
        if inImport {
            if strings.HasPrefix(s, "-") {
                item := strings.TrimSpace(strings.TrimPrefix(s, "-"))
                // Only keep the first token (path); ignore constraints
                parts := strings.Fields(item)
                if len(parts) >= 1 {
                    imports = append(imports, parts[0])
                }
            } else if s == "" || strings.HasPrefix(s, "#") {
                // continue
            } else {
                inImport = false
            }
        }
    }
    return imports, nil
}

// isDeclaredLocalImport returns true if rel path ("./subproject") is present in any import list.
func isDeclaredLocalImport(wsPath, rel string) (bool, error) {
    imports, err := parseWorkspaceImports(wsPath)
    if err != nil { return false, err }
    // Normalize forms to compare: allow with or without leading ./
    cand := rel
    if !strings.HasPrefix(cand, "./") {
        cand = "./" + strings.TrimPrefix(cand, "./")
    }
    for _, imp := range imports {
        i := imp
        if !strings.HasPrefix(i, "./") { i = "./" + strings.TrimPrefix(i, "./") }
        if i == cand { return true, nil }
    }
    return false, nil
}

