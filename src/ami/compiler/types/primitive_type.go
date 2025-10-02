package types

type Primitive struct{ K Kind }

func (p Primitive) String() string { return p.K.String() }

