package main

import (
    "io"
    "github.com/sam-caldwell/ami/src/ami/logging"
)

var rootLogger *logging.Logger

func getRootLogger() *logging.Logger { return rootLogger }

// discardWriter is exposed for testing/internals.
var discardWriter io.Writer
