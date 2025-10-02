package sem

// decorator disable set (scaffold for workspace wiring)
var disabledDecorators = map[string]struct{}{}

// SetDisabledDecorators replaces the set of disabled decorator names for analysis.
// Names are matched case-sensitively against full name and last-segment base.
func SetDisabledDecorators(names ...string) {
    m := map[string]struct{}{}
    for _, n := range names { if n != "" { m[n] = struct{}{} } }
    disabledDecorators = m
}

// AnalyzeDecorators performs basic resolution and consistency checks for function decorators.
// Scaffold rules:
// - Built-ins: deprecated, metrics (recognized without requiring a top-level symbol)
// - Otherwise, decorator name must resolve to a top-level function declared in the same file
// - Duplicate decorator with different arg list emits E_DECORATOR_CONFLICT
// - Unresolved decorator emits E_DECORATOR_UNRESOLVED
