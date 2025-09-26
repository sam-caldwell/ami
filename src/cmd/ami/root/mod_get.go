package root

import (
    ammod "github.com/sam-caldwell/ami/src/ami/mod"
    ex "github.com/sam-caldwell/ami/src/internal/exit"
    "github.com/sam-caldwell/ami/src/internal/logger"
    "github.com/spf13/cobra"
    "os"
    "strings"
)

func newModGetCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "get <url>",
        Short: "Fetch a package into the cache",
        Args:  cobra.MinimumNArgs(1),
        Example: `  ami mod get ./local/module
  ami mod get git+ssh://git@github.com/org/repo.git#v1.2.3`,
        Run: func(cmd *cobra.Command, args []string) {
            u := args[0]
            dest, pkg, ver, err := ammod.GetWithInfo(u)
            if err != nil {
                if strings.HasPrefix(u, "git+ssh://") {
                    // Treat any git+ssh failure as a network/registry error for this phase
                    logger.Error("network registry error", map[string]interface{}{"url": u, "error": err.Error()})
                    os.Stderr.WriteString("network registry error\n")
                    os.Exit(ex.NetworkRegistryError)
                }
                logger.Error(err.Error(), map[string]interface{}{"url": u})
                return
            }
            // If the backend provided package/version info (e.g., git+ssh), update ami.sum best-effort
            if pkg != "" && ver != "" {
                if uerr := ammod.UpdateSum("ami.sum", pkg, ver, dest, ver); uerr != nil {
                    logger.Warn("failed to update ami.sum", map[string]interface{}{"error": uerr.Error()})
                }
            }
            logger.Info("fetched", map[string]interface{}{"dest": dest})
        },
    }
}

