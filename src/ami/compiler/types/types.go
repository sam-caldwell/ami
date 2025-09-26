package types

// Type is the common interface implemented by all AMI types.
type Type interface{ String() string }

// Basic represents a simple named type (e.g., int, string).
type Basic struct{ name string }

func (b Basic) String() string { return b.name }

var (
	TInvalid = Basic{"invalid"}
	TInt     = Basic{"int"}
	TFloat   = Basic{"float"}
	TString  = Basic{"string"}
	TBool    = Basic{"bool"}
	TPackage = Basic{"package"}
)
