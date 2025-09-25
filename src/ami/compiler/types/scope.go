package types

import "fmt"

// Scope stores objects with lexical scoping and stable iteration order.
type Scope struct {
	parent *Scope
	order  []string
	table  map[string]*Object
}

func NewScope(parent *Scope) *Scope { return &Scope{parent: parent, table: make(map[string]*Object)} }

// Insert adds an object; returns error if already declared in this scope.
func (s *Scope) Insert(obj *Object) error {
	if _, exists := s.table[obj.Name]; exists {
		return fmt.Errorf("duplicate: %s", obj.Name)
	}
	s.table[obj.Name] = obj
	s.order = append(s.order, obj.Name)
	return nil
}

// Lookup searches this scope, then parents.
func (s *Scope) Lookup(name string) *Object {
	if obj, ok := s.table[name]; ok {
		return obj
	}
	if s.parent != nil {
		return s.parent.Lookup(name)
	}
	return nil
}

// Names returns declared names in insertion order.
func (s *Scope) Names() []string {
	out := make([]string, len(s.order))
	copy(out, s.order)
	return out
}
