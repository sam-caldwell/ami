package workspace

// FindPackage returns a pointer to the package with the given key, if present.
func (w *Workspace) FindPackage(key string) *Package {
    for i := range w.Packages {
        if w.Packages[i].Key == key {
            return &w.Packages[i].Package
        }
    }
    return nil
}

