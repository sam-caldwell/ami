package main

import (
    "io"
    "github.com/sam-caldwell/ami/src/ami/logging"
)

var rootLogger *logging.Logger

func setRootLogger(l *logging.Logger) { rootLogger = l }

func getRootLogger() *logging.Logger { return rootLogger }

// closeRootLogger closes the global logger if present.
func closeRootLogger() {
    if rootLogger != nil {
        _ = rootLogger.Close()
        // drain reference so tests can verify fresh state across runs
        rootLogger = nil
    }
}

// discardWriter is exposed for testing/internals.
var discardWriter io.Writer

