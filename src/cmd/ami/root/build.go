package root

import (
    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/internal/logger"
)

var cmdBuild = &cobra.Command{
    Use:   "build",
    Short: "Build the workspace",
    Run: func(cmd *cobra.Command, args []string) {
        logger.Info("ami build: not yet implemented", nil)
    },
}

