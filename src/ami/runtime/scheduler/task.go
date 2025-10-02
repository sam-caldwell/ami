package scheduler

import "context"

// Task represents a unit of work.
type Task struct {
    Source string
    Do     func(context.Context)
}

