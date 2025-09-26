package root

import (
    "errors"
    ammod "github.com/sam-caldwell/ami/src/ami/mod"
    ex "github.com/sam-caldwell/ami/src/internal/exit"
    "github.com/sam-caldwell/ami/src/internal/logger"
    "github.com/spf13/cobra"
    "os"
)

func newModUpdateCmd() *cobra.Command {
    return &cobra.Command{
        Use:     "update",
        Short:   "Update project dependencies",
        Example: `  ami mod update`,
        Run: func(cmd *cobra.Command, args []string) {
            if err := ammod.UpdateFromWorkspace("ami.workspace"); err != nil {
                if errors.Is(err, ammod.ErrNetwork) {
                    logger.Error("network registry error", map[string]interface{}{"error": err.Error()})
                    os.Stderr.WriteString("network registry error\n")
                    os.Exit(ex.NetworkRegistryError)
                }
                logger.Error(err.Error(), nil)
                return
            }
            logger.Info("dependencies updated (ami.sum)", nil)
        },
    }
}

