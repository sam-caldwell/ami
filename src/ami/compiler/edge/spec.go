package edge

// Spec is the common interface for edge specifications.
type Spec interface {
    Kind() Kind
    // Validate performs minimal structural validation of the spec.
    Validate() error
}

