package main

import (
	"github.com/sam-caldwell/ami/src/cmd/ami/root"
	"os"
)

func main() {
	code := root.Execute()
	os.Exit(code)
}
