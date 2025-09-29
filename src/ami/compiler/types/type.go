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

// Named represents a named type or type variable (e.g., user type or single-letter T).
type Named struct{ Name string }

func (n Named) String() string { return n.Name }

// Optional wraps a single inner type representing an optional value.
type Optional struct{ Inner Type }

func (o Optional) String() string { return "Optional<" + o.Inner.String() + ">" }

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

// Equal reports structural equality of two types, ignoring Union alternative
// order. Struct field order is normalized by String() so direct comparison of
// keys with recursive Equal suffices.
func Equal(a, b Type) bool {
    switch av := a.(type) {
    case Primitive:
        bv, ok := b.(Primitive); return ok && av.K == bv.K
    case Named:
        bv, ok := b.(Named); return ok && av.Name == bv.Name
    case Generic:
        bv, ok := b.(Generic); if !ok { return false }
        if av.Name != bv.Name || len(av.Args) != len(bv.Args) { return false }
        // Special-case Union represented as Generic shouldn't occur; Union has its own type.
        for i := range av.Args { if !Equal(av.Args[i], bv.Args[i]) { return false } }
        return true
    case Struct:
        bv, ok := b.(Struct); if !ok { return false }
        if len(av.Fields) != len(bv.Fields) { return false }
        for k, v := range av.Fields {
            vv, ok := bv.Fields[k]; if !ok { return false }
            if !Equal(v, vv) { return false }
        }
        return true
    case Optional:
        bv, ok := b.(Optional); return ok && Equal(av.Inner, bv.Inner)
    case Union:
        bv, ok := b.(Union); if !ok { return false }
        if len(av.Alts) != len(bv.Alts) { return false }
        // Compare as sets using string keys built from structural String forms.
        set := make(map[string]struct{}, len(av.Alts))
        for _, x := range av.Alts { set[x.String()] = struct{}{} }
        for _, y := range bv.Alts { if _, ok := set[y.String()]; !ok { return false } }
        return true
    default:
        // Unknown concrete type; fall back to string comparison as last resort
        return a != nil && b != nil && a.String() == b.String()
    }
}
