package merge

import (
    "context"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
    "github.com/sam-caldwell/ami/src/ami/runtime/scheduler"
)

// MergeTask creates a scheduler.Task that runs the merge RunPlan loop.
// Source labels the task for FAIR scheduling.
func MergeTask(source string, plan Plan, in <-chan ev.Event, out chan<- ev.Event) scheduler.Task {
    return scheduler.Task{Source: source, Do: func(ctx context.Context){ RunPlan(ctx, plan, in, out) }}
}

