package memory

// NewManager creates a new memory Manager with zeroed counters.
func NewManager() *Manager {
    return &Manager{counters: map[Domain]int{Event: 0, State: 0, Ephemeral: 0}}
}

