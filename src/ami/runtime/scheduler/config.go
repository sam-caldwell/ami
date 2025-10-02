package scheduler

// Config controls worker model per node kind.
type Config struct {
    Workers       int
    Policy        Policy
    QueueCapacity int // 0=unbounded in practice
}

