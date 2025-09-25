package root

import (
    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/internal/logger"
)

var cmdClean = &cobra.Command{
    Use:   "clean",
    Short: "Clean build artifacts",
    Run: func(cmd *cobra.Command, args []string) {
        logger.Info("ami clean: not yet implemented", nil)
    },
}

