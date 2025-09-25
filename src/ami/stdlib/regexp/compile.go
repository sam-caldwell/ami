package regexp

import stdregexp "regexp"

// Compile parses a regular expression and returns, if successful, a Regexp object.
func Compile(pattern string) (*Regexp, error) {
    re, err := stdregexp.Compile(pattern)
    if err != nil { return nil, err }
    return &Regexp{re: re}, nil
}

