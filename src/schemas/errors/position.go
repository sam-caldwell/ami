package errors

// Position mirrors diag.Position for schema alignment and stability.
type Position struct {
    Line   int `json:"line"`
    Column int `json:"column"`
    Offset int `json:"offset"`
}

