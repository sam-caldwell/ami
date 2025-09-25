package root

import (
    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/internal/logger"
)

var cmdLint = &cobra.Command{
    Use:   "lint",
    Short: "Lint the project",
    Run: func(cmd *cobra.Command, args []string) {
        logger.Info("ami lint: not yet implemented", nil)
    },
}

