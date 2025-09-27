package types

// Map represents a map generic form (semantic form "map<K,V>").
type Map struct{ Key, Val Type }

func (m Map) String() string { return "map<" + m.Key.String() + "," + m.Val.String() + ">" }

