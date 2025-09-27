package types

import "strings"

// String renders a stable textual representation of the function type.
func (f Function) String() string {
    ps := make([]string, len(f.Params))
    for i, p := range f.Params { ps[i] = p.String() }
    rs := make([]string, len(f.Results))
    for i, r := range f.Results { rs[i] = r.String() }
    return "func(" + strings.Join(ps, ",") + ") -> (" + strings.Join(rs, ",") + ")"
}

