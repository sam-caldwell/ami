package regexp

// MatchString reports whether the pattern matches the string s.
func (r *Regexp) MatchString(s string) bool { return r.re.MatchString(s) }

