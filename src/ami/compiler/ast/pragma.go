package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// Pragma captures a single '#pragma ...' directive as raw text with a position.
// Parsing/interpretation is deferred to consumers (e.g., test harness or linter).
type Pragma struct {
    Pos  source.Position
    Text string
    Domain string            // e.g., "lint", "test"
    Key    string            // e.g., "disable", "case"
    Value  string            // unparsed remainder after domain:key
    Args   []string          // space-separated args (raw tokens)
    Params map[string]string // tokens like key=value parsed into map
}
