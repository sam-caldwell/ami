package root

import (
    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/internal/logger"
)

func newLintCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "lint",
        Short: "Lint the project",
        Example: `  ami lint
  ami --json lint`,
        Run: func(cmd *cobra.Command, args []string) {
            logger.Info("ami lint: not yet implemented", nil)
        },
    }
}
