package types

// Type is a marker interface for AMI types.
type Type interface{ String() string }

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
