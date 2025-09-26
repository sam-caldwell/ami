package schemas

// EdgesV1 summarizes compiler-discovered input edges per package
// for debug/inspection. Emitted under build/debug/asm/<package>/edges.json
// when `ami build --verbose` is used.
type EdgesV1 struct {
    Schema    string       `json:"schema"`
    Timestamp string       `json:"timestamp"`
    Package   string       `json:"package"`
    Items     []EdgeInitV1 `json:"items"`
}

