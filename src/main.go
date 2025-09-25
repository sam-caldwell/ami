package main

import (
    "os"
    root "github.com/sam-caldwell/ami/src/cmd/ami/root"
)

func main() {
    os.Exit(root.Execute())
}
