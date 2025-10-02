package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/logging"
)

func Test_setRootLogger_exists(t *testing.T) {
    setRootLogger((*logging.Logger)(nil))
}

