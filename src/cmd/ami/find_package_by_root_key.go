package main

import "github.com/sam-caldwell/ami/src/ami/workspace"

// findPackageByRootKey returns a package by the PackageList key (e.g., "main").
func findPackageByRootKey(ws *workspace.Workspace, key string) *workspace.Package {
    for i := range ws.Packages { if ws.Packages[i].Key == key { return &ws.Packages[i].Package } }
    return nil
}

