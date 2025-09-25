package main

import (
	root "github.com/sam-caldwell/ami/src/cmd/ami/root"
	"os"
)

func main() {
	os.Exit(root.Execute())
}
