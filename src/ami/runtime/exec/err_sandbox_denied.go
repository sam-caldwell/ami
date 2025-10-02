package exec

// ErrSandboxDenied is returned when a capability is denied by the sandbox policy.
type ErrSandboxDenied struct{ Cap string }

func (e ErrSandboxDenied) Error() string { return "sandbox denied capability: " + e.Cap }

