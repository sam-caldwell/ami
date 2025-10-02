package main

import (
    "fmt"
    "os"
    "github.com/sam-caldwell/ami/src/ami/exit"
)

// execute runs the root command and returns an exit code.
func execute() int {
    root := newRootCmd()
    if err := root.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err.Error())
        code := exit.UnwrapCode(err)
        if code == exit.Internal { code = exit.User }
        return code.Int()
    }
    return 0
}

