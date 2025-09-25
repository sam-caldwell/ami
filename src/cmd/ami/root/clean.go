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
        _ = os.RemoveAll("build")
        if err := os.MkdirAll("build", 0755); err != nil {
            logger.Error("failed to recreate build directory", map[string]interface{}{"error": err.Error()})
            return
        }
        logger.Info("cleaned build directory", map[string]interface{}{"path": "build"})
    },
}

