package root

import (
	"fmt"
	"github.com/sam-caldwell/ami/src/internal/logger"
	"github.com/spf13/cobra"
)

var version = "v0.0.0-dev" // overridden via -ldflags at build time

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Example: `  ami version
  ami --json version`,
		Run: func(cmd *cobra.Command, args []string) {
			if flagJSON {
				fmt.Printf("{\"version\":\"%s\"}\n", version)
				return
			}
			logger.Info(fmt.Sprintf("version: %s", version), map[string]interface{}{"version": version})
		},
	}
}
