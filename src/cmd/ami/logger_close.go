package main

// closeRootLogger closes the global logger if present.
func closeRootLogger() {
    if rootLogger != nil {
        _ = rootLogger.Close()
        // drain reference so tests can verify fresh state across runs
        rootLogger = nil
    }
}

