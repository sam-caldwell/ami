package types

import astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// FromAST maps an AST TypeRef into a types.Type.
func FromAST(tr astpkg.TypeRef) Type {
	// Base type name (case-sensitive as in AST)
	base := tr.Name
	// Map generics first
	switch base {
	case "Event":
		if len(tr.Args) == 1 {
			return EventType{Elem: FromAST(tr.Args[0])}
		}
		return TInvalid
	case "Error":
		if len(tr.Args) == 1 {
			return ErrorType{Elem: FromAST(tr.Args[0])}
		}
		return TInvalid
	case "Owned":
		if len(tr.Args) == 1 {
			return OwnedType{Elem: FromAST(tr.Args[0])}
		}
		return TInvalid
	case "map":
		if len(tr.Args) == 2 {
			return Map{Key: FromAST(tr.Args[0]), Value: FromAST(tr.Args[1])}
		}
		return TInvalid
	case "set":
		if len(tr.Args) == 1 {
			return Set{Elem: FromAST(tr.Args[0])}
		}
		return TInvalid
	case "slice":
		if len(tr.Args) == 1 {
			return SliceTy{Elem: FromAST(tr.Args[0])}
		}
		return TInvalid
	}
	// Basic types (subset)
	switch base {
	case "int":
		return TInt
	case "float":
		return TFloat
	case "string":
		return TString
	case "bool":
		return TBool
	}
	// Named types
	t := Basic{name: base}
	if tr.Ptr {
		return Pointer{Base: t}
	}
	if tr.Slice {
		return Slice{Elem: t}
	}
	// if slice ptr/generics already handled above
	return t
}
