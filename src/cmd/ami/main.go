package main

import (
    "os"
    "github.com/sam-caldwell/ami/src/cmd/ami/root"
)

func main() {
    code := root.Execute()
    os.Exit(code)
}

