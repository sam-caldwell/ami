package root

import (
    "os"
    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/internal/logger"
)

var cmdClean = &cobra.Command{
    Use:   "clean",
    Short: "Clean build artifacts",
    Run: func(cmd *cobra.Command, args []string) {
        // Remove build directory if present
        if _, err := os.Stat("build"); err == nil {
            if err := os.RemoveAll("build"); err != nil {
                logger.Error("clean.remove_failed", map[string]interface{}{"path": "build", "error": err.Error()})
                return
            }
            logger.Info("clean.remove", map[string]interface{}{"path": "build"})
        } else {
            logger.Info("clean.remove.skip", map[string]interface{}{"path": "build", "reason": "not found"})
        }
        // Recreate build directory
        if err := os.MkdirAll("build", 0755); err != nil {
            logger.Error("clean.create_failed", map[string]interface{}{"path": "build", "error": err.Error()})
            return
        }
        logger.Info("clean.create", map[string]interface{}{"path": "build"})
    },
}
