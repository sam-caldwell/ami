package root

import (
	ammod "github.com/sam-caldwell/ami/src/ami/mod"
	"github.com/sam-caldwell/ami/src/internal/logger"
	"github.com/spf13/cobra"
	"net/url"
	"path/filepath"
	"strings"
)

var cmdMod = &cobra.Command{
	Use:   "mod",
	Short: "Module and cache operations",
}

var cmdModClean = &cobra.Command{
	Use:   "clean",
	Short: "Clean the module cache",
	Run: func(cmd *cobra.Command, args []string) {
		dir, err := ammod.CacheDir()
		if err != nil {
			logger.Error(err.Error(), nil)
			return
		}
		logger.Info("cache directory", map[string]interface{}{"path": dir})
	},
}

var cmdModUpdate = &cobra.Command{
	Use:   "update",
	Short: "Update project dependencies",
	Run:   func(cmd *cobra.Command, args []string) { logger.Info("ami mod update: not yet implemented", nil) },
}

var cmdModGet = &cobra.Command{
	Use:   "get <url>",
	Short: "Fetch a package into the cache",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dest, err := ammod.Get(args[0])
		if err != nil {
			logger.Error(err.Error(), map[string]interface{}{"url": args[0]})
			return
		}
		logger.Info("fetched", map[string]interface{}{"dest": dest})
	},
}

var cmdModList = &cobra.Command{
	Use:   "list",
	Short: "List cached packages",
	Run: func(cmd *cobra.Command, args []string) {
		items, err := ammod.List()
		if err != nil {
			logger.Error(err.Error(), nil)
			return
		}
		for _, it := range items {
			logger.Info(it, nil)
		}
	},
}

func init() {
	cmdMod.AddCommand(cmdModClean)
	cmdMod.AddCommand(cmdModUpdate)
	cmdMod.AddCommand(cmdModGet)
	cmdMod.AddCommand(cmdModList)
	cmdMod.AddCommand(cmdModVerify)
}

var cmdModVerify = &cobra.Command{
	Use:   "verify",
	Short: "Verify ami.sum against cache",
	Run: func(cmd *cobra.Command, args []string) {
		// Simple verification: ensure each cached entry exists; if git repo, re-compute digest
		cacheDir, err := ammod.CacheDir()
		if err != nil {
			logger.Error(err.Error(), nil)
			return
		}
		sum, err := ammod.LoadSumForCLI("ami.sum")
		if err != nil {
			logger.Error(err.Error(), nil)
			return
		}
		ok := true
		_ = cacheDir
		_ = sum
		_ = ok
		logger.Info("ami mod verify: not yet implemented", nil)
	},
}
