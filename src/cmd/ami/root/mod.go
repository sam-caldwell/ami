package root

import (
    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/internal/logger"
)

var cmdMod = &cobra.Command{
    Use:   "mod",
    Short: "Module and cache operations",
}

var cmdModClean = &cobra.Command{
    Use:   "clean",
    Short: "Clean the module cache",
    Run: func(cmd *cobra.Command, args []string) { logger.Info("ami mod clean: not yet implemented", nil) },
}

var cmdModUpdate = &cobra.Command{
    Use:   "update",
    Short: "Update project dependencies",
    Run: func(cmd *cobra.Command, args []string) { logger.Info("ami mod update: not yet implemented", nil) },
}

var cmdModGet = &cobra.Command{
    Use:   "get <url>",
    Short: "Fetch a package into the cache",
    Args:  cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) { logger.Info("ami mod get: not yet implemented", map[string]interface{}{"url": args[0]}) },
}

var cmdModList = &cobra.Command{
    Use:   "list",
    Short: "List cached packages",
    Run: func(cmd *cobra.Command, args []string) { logger.Info("ami mod list: not yet implemented", nil) },
}

func init() {
    cmdMod.AddCommand(cmdModClean)
    cmdMod.AddCommand(cmdModUpdate)
    cmdMod.AddCommand(cmdModGet)
    cmdMod.AddCommand(cmdModList)
}

