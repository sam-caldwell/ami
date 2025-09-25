package root

import (
    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/internal/logger"
)

var cmdTest = &cobra.Command{
    Use:   "test",
    Short: "Run tests",
    Run: func(cmd *cobra.Command, args []string) {
        logger.Info("ami test: not yet implemented", nil)
    },
}

