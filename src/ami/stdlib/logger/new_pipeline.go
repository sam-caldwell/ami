package logger

// NewPipeline creates an unstarted pipeline with the given config.
func NewPipeline(cfg Config) *Pipeline {
    cap := cfg.Capacity
    if cap < 0 { cap = 0 }
    p := &Pipeline{
        cfg:  cfg,
        ch:   make(chan []byte, cap),
        stop: make(chan struct{}),
    }
    return p
}

