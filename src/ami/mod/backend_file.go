package mod

import (
    "errors"
    "os"
    "path/filepath"
    "strings"
)

type fileBackend struct{}

func (fileBackend) Name() string { return "file" }

func (fileBackend) Match(spec string) bool {
    return strings.HasPrefix(spec, "file://") || strings.HasPrefix(spec, "./") || strings.HasPrefix(spec, "../") || strings.HasPrefix(spec, "/")
}

func (fileBackend) Fetch(spec, cacheDir string) (string, string, string, error) {
    // Support optional file:// prefix; strip it for path operations
    path := spec
    if strings.HasPrefix(path, "file://") {
        path = strings.TrimPrefix(path, "file://")
        // For consistency, a lone host part is not expected; treat as local filesystem path
    }
    // Resolve workspace and validate declaration
    cwd, _ := os.Getwd()
    wsRoot, err := findWorkspaceRoot(cwd)
    if err != nil { return "", "", "", err }
    absSrc, err := filepath.Abs(path)
    if err != nil { return "", "", "", err }
    absRoot, _ := filepath.Abs(wsRoot)
    relToRoot, err := filepath.Rel(absRoot, absSrc)
    if err != nil { return "", "", "", err }
    if strings.HasPrefix(relToRoot, "..") { return "", "", "", errors.New("local path must be within workspace") }
    wsPath := filepath.Join(wsRoot, "ami.workspace")
    relDecl := filepath.ToSlash("./" + relToRoot)
    ok, err := isDeclaredLocalImport(wsPath, relDecl)
    if err != nil { return "", "", "", err }
    if !ok { return "", "", "", errors.New("local path not declared in ami.workspace imports") }
    // Copy into cache under <name>@local
    name := filepath.Base(absSrc)
    dest := filepath.Join(cacheDir, name+"@local")
    _ = os.RemoveAll(dest)
    if err := copyDir(absSrc, dest); err != nil { return "", "", "", err }
    // File backend does not participate in ami.sum directly
    return dest, "", "", nil
}

func init() { registerBackend(fileBackend{}) }

