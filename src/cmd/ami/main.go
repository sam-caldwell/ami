package main

import (
    "os"
)

// main is the entrypoint for the ami CLI.
func main() {
    // Defer to the root command built in root.go. We avoid printing
    // anything here to keep CLI deterministic and testable.
    code := execute()
    // Close the root logger if present before exiting.
    closeRootLogger()
    os.Exit(code)
}
