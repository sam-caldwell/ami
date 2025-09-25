package regexp

import stdregexp "regexp"

// Regexp wraps Go's RE2-based regexp for deterministic matching.
type Regexp struct{ re *stdregexp.Regexp }
