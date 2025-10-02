package types

import "strings"

// Union represents a set of alternative types. Order is not semantically
// important; String() prints a stable, sorted representation for determinism.
type Union struct{ Alts []Type }

func (u Union) String() string {
    if len(u.Alts) == 0 { return "Union<>" }
    // derive stable order by string representations
    parts := make([]string, len(u.Alts))
    for i, a := range u.Alts { parts[i] = a.String() }
    // simple insertion sort to avoid importing sort unnecessarily here
    for i := 1; i < len(parts); i++ {
        for j := i; j > 0 && parts[j] < parts[j-1]; j-- {
            parts[j], parts[j-1] = parts[j-1], parts[j]
        }
    }
    return "Union<" + strings.Join(parts, ",") + ">"
}

