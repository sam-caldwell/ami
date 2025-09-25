package root

import (
	"github.com/sam-caldwell/ami/src/internal/logger"
	"github.com/spf13/cobra"
	"os"
)

func hasAnyWritePermission(dir string) bool {
	fi, err := os.Stat(dir)
	if err != nil {
		return false
	}
	// coarse check: any of user/group/other write bits
	return fi.Mode().Perm()&0o222 != 0
}

func newCleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Clean build artifacts",
		Example: `  # Clean build directory
  ami clean

  # Emit JSON records
  ami --json clean`,
		Run: func(cmd *cobra.Command, args []string) {
			// Remove build directory if present
			if _, err := os.Stat("build"); err == nil {
				if err := os.RemoveAll("build"); err != nil || !hasAnyWritePermission(".") {
					if err != nil {
						logger.Error("clean.remove_failed", map[string]interface{}{"path": "build", "error": err.Error()})
					} else {
						logger.Error("clean.remove_failed", map[string]interface{}{"path": "build", "error": "no write permission"})
					}
					return
				}
				logger.Info("clean.remove", map[string]interface{}{"path": "build"})
			} else {
				logger.Info("clean.remove.skip", map[string]interface{}{"path": "build", "reason": "not found"})
			}
			// Recreate build directory
			if err := os.MkdirAll("build", 0755); err != nil || !hasAnyWritePermission(".") {
				if err != nil {
					logger.Error("clean.create_failed", map[string]interface{}{"path": "build", "error": err.Error()})
				} else {
					logger.Error("clean.create_failed", map[string]interface{}{"path": "build", "error": "no write permission"})
				}
				return
			}
			logger.Info("clean.create", map[string]interface{}{"path": "build"})
		},
	}
}
