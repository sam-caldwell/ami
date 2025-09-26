package memory

// Domain identifies an allocation domain in the AMI memory model.
// Event: per-event heap; State: node-state heap; Ephemeral: call-local.
type Domain string

const (
    Event     Domain = "event"
    State     Domain = "state"
    Ephemeral Domain = "ephemeral"
)

