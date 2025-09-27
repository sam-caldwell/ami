package diag

// Position carries source location information for diagnostics.
type Position struct {
    Line   int `json:"line"`
    Column int `json:"column"`
    Offset int `json:"offset"`
}

