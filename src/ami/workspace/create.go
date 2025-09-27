package workspace

// Create writes a default workspace to the provided path.
func (w *Workspace) Create(path string) error {
    *w = DefaultWorkspace()
    return w.Save(path)
}

