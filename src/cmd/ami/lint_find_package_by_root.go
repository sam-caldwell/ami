package main

import "github.com/sam-caldwell/ami/src/ami/workspace"

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

