package root

import (
    ammod "github.com/sam-caldwell/ami/src/ami/mod"
    "github.com/sam-caldwell/ami/src/internal/logger"
    "github.com/spf13/cobra"
    "os"
)

func newModCleanCmd() *cobra.Command {
    return &cobra.Command{
        Use:     "clean",
        Short:   "Clean the module cache",
        Example: `  ami mod clean`,
        Run: func(cmd *cobra.Command, args []string) {
            // Determine cache dir path without creating it
            dir, err := ammod.CacheDirPath()
            if err != nil { logger.Error(err.Error(), nil); return }
            if _, statErr := os.Stat(dir); statErr == nil {
                if err := os.RemoveAll(dir); err != nil {
                    logger.Error("cache.remove_failed", map[string]interface{}{"path": dir, "error": err.Error()})
                    return
                }
                logger.Info("cache.remove", map[string]interface{}{"path": dir})
            } else {
                logger.Info("cache.remove.skip", map[string]interface{}{"path": dir, "reason": "not found"})
            }
            // Recreate cache directory
            if err := os.MkdirAll(dir, 0o755); err != nil {
                logger.Error("cache.create_failed", map[string]interface{}{"path": dir, "error": err.Error()})
                return
            }
            logger.Info("cache.create", map[string]interface{}{"path": dir})
        },
    }
}

