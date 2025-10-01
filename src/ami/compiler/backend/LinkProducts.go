package backend

// LinkProducts captures outputs of the linker.
type LinkProducts struct {
	// Binary is the path to the produced executable (workspace-relative).
	Binary string
}
