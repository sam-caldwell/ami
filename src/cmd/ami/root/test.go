package root

import (
    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/internal/logger"
)

func newTestCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "test",
        Short: "Run tests",
        Example: `  ami test
  ami --json test`,
        Run: func(cmd *cobra.Command, args []string) {
            logger.Info("ami test: not yet implemented", nil)
        },
    }
}
