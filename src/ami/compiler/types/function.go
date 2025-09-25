package types

// Function represents a function signature with ordered params and results.
type Function struct {
	Params  []Type
	Results []Type
}

func (f Function) String() string {
	// minimal rendering: (p1,p2)->(r1,r2)
	s := "("
	for i, p := range f.Params {
		if i > 0 {
			s += ","
		}
		s += p.String()
	}
	s += ") -> ("
	for i, r := range f.Results {
		if i > 0 {
			s += ","
		}
		s += r.String()
	}
	s += ")"
	return s
}
