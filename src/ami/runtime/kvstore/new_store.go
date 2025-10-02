package kvstore

import "container/list"

// New creates a new empty Store.
func New() *Store { return &Store{items: make(map[string]*entry), lru: list.New(), order: map[string]*list.Element{}} }

