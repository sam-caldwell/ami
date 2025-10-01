package backend

// Linker is implemented by backends that can link objects for a target.
type Linker interface {
	Link(opts LinkOptions) (LinkProducts, error)
}
