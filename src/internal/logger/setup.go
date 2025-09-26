package logger

// Setup configures the global logger mode.
func Setup(jsonMode, verbose, color bool) {
    std.mu.Lock()
    defer std.mu.Unlock()
    std.json = jsonMode
    // If JSON, force color off
    if jsonMode {
        color = false
    }
    std.verbose = verbose
    std.color = color
}

