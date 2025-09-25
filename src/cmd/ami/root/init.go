package root

import (
    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/internal/logger"
)

var cmdInit = &cobra.Command{
    Use:   "init",
    Short: "Initialize the AMI workspace",
    Run: func(cmd *cobra.Command, args []string) {
        logger.Info("ami init: not yet implemented", nil)
    },
}

