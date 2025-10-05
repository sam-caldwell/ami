package exec

import ev "github.com/sam-caldwell/ami/src/schemas/events"

// WorkerInvoker resolves and invokes compiled worker symbols.
//
// Implementations may use dynamic linking (e.g., dlsym) to resolve symbols
// following a stable ABI. The resolved function returns either an events.Event
// (forwarded unchanged) or a bare payload value to be wrapped into the current
// event by the caller. Errors indicate worker failure and should be routed to
// the error channel when configured.
type WorkerInvoker interface {
    Resolve(workerName string) (func(ev.Event) (any, error), bool)
}

