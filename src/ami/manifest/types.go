package manifest

// Manifest represents the ami.manifest document in memory.
// It stores the declared schema version and an open-ended data map
// for forward-compatible fields produced/consumed by other packages.
//
// The Save/Load semantics aim to be deterministic at the top level:
// - "schema" is always written first.
// - remaining top-level keys from Data are emitted in lexicographic order.
type Manifest struct {
    // Schema is the document schema identifier (e.g., "ami.manifest/v1").
    Schema string
    // Data carries all additional top-level fields in the manifest.
    Data map[string]any
}

