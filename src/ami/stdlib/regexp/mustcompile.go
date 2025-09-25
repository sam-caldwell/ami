package regexp

import stdregexp "regexp"

// MustCompile is like Compile but panics if the expression cannot be parsed.
func MustCompile(pattern string) *Regexp { return &Regexp{re: stdregexp.MustCompile(pattern)} }

