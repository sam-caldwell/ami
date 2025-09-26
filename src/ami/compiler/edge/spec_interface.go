package edge

// Spec is the common interface for all edge declarations.
// Implementations validate their parameters but do not perform runtime I/O.
type Spec interface {
    Kind() string
    Validate() error
}

