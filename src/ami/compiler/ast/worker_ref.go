package ast

// WorkerRef references a worker by name and indicates its kind.
// Kind values include "function" and "factory".
type WorkerRef struct {
    Name string
    Kind string
}

