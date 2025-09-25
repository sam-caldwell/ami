package root

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/internal/logger"
)

var version = "v0.0.0-dev" // overridden via -ldflags at build time

var cmdVersion = &cobra.Command{
    Use:   "version",
    Short: "Print version information",
    Run: func(cmd *cobra.Command, args []string) {
        if flagJSON {
            fmt.Printf("{\"version\":\"%s\"}\n", version)
            return
        }
        logger.Info(fmt.Sprintf("version: %s", version), map[string]interface{}{"version": version})
    },
}
