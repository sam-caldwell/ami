package types

import "strings"

// Type is implemented by all types.
type Type interface{ String() string }

// Primitive represents a builtin type.
type Primitive struct{ K Kind }

func (p Primitive) String() string { return p.K.String() }

// Generic represents parameterized types like Event<T>, Error<E>, Owned<T>.
type Generic struct {
    Name string
    Args []Type
}

func (g Generic) String() string {
    if len(g.Args) == 0 { return g.Name }
    parts := make([]string, len(g.Args))
    for i, a := range g.Args { parts[i] = a.String() }
    return g.Name + "<" + strings.Join(parts, ",") + ">"
}

