package root

import (
    manifest "github.com/sam-caldwell/ami/src/ami/manifest"
    ammod "github.com/sam-caldwell/ami/src/ami/mod"
    "github.com/sam-caldwell/ami/src/internal/logger"
    "github.com/spf13/cobra"
    "net/url"
    "os"
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

var cmdModUpdate = &cobra.Command{
	Use:   "update",
	Short: "Update project dependencies",
	Run:   func(cmd *cobra.Command, args []string) {
        if err := ammod.UpdateFromWorkspace("ami.workspace"); err != nil {
            logger.Error(err.Error(), nil)
            return
        }
        logger.Info("dependencies updated (ami.sum)", nil)
    },
}

var cmdModGet = &cobra.Command{
	Use:   "get <url>",
	Short: "Fetch a package into the cache",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		u := args[0]
		dest, err := ammod.Get(u)
		if err != nil {
			logger.Error(err.Error(), map[string]interface{}{"url": u})
			return
		}

		if strings.HasPrefix(u, "git+ssh://") {
			raw := strings.TrimPrefix(u, "git+")
			tag := ""
			if i := strings.Index(raw, "#"); i >= 0 {
				tag = raw[i+1:]
				raw = raw[:i]
			}
			if tag != "" {
				if parsed, perr := url.Parse(raw); perr == nil {
					host := parsed.Host
					repoPath := strings.TrimSuffix(strings.TrimPrefix(parsed.Path, "/"), ".git")
					pkg := filepath.Join(host, repoPath)
					if uerr := ammod.UpdateSum("ami.sum", pkg, tag, dest, tag); uerr != nil {
						logger.Warn("failed to update ami.sum", map[string]interface{}{"error": uerr.Error()})
					}
				}
			}
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
        // Attempt to load ami.sum for digest lookup (best-effort)
        sum, _ := ammod.LoadSumForCLI("ami.sum")
        // Attempt to load ami.manifest for richer package info (best-effort)
        var manPkgs map[string]map[string]struct{ Full string; Digest string; Source string }
        if m, err := manifest.Load("ami.manifest"); err == nil {
            manPkgs = make(map[string]map[string]struct{ Full string; Digest string; Source string })
            for _, p := range m.Packages {
                base := filepath.Base(p.Name)
                if _, ok := manPkgs[base]; !ok { manPkgs[base] = make(map[string]struct{ Full string; Digest string; Source string }) }
                manPkgs[base][p.Version] = struct{ Full string; Digest string; Source string }{ Full: p.Name, Digest: p.Digest, Source: p.Source }
            }
        }
        for _, it := range items {
            if flagJSON {
                // it format: <name>@<version>
                name := it
                ver := ""
                if i := strings.LastIndex(it, "@"); i >= 0 {
                    name = it[:i]
                    ver = it[i+1:]
                }
                // Find digest by matching base name against sum packages
                var digest string
                if sum != nil {
                    for pkg, vers := range sum.Packages {
                        if filepath.Base(pkg) == name {
                            if d, ok := vers[ver]; ok {
                                digest = d
                                break
                            }
                        }
                    }
                }
                data := map[string]interface{}{
                    "entry":   it,
                    "name":    name,
                    "version": ver,
                }
                if digest != "" {
                    data["digest"] = digest
                }
                // Augment with manifest info if available
                if manPkgs != nil {
                    if versions, ok := manPkgs[name]; ok {
                        if info, ok2 := versions[ver]; ok2 {
                            data["manifest"] = true
                            data["manifestName"] = info.Full
                            if info.Digest != "" { data["manifestDigest"] = info.Digest }
                            if info.Source != "" { data["manifestSource"] = info.Source }
                        }
                    }
                }
                logger.Info("cache.entry", data)
            } else {
                logger.Info(it, nil)
            }
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
        for pkg, vers := range sum.Packages {
            base := filepath.Base(pkg)
            for ver, digest := range vers {
                entry := filepath.Join(cacheDir, base+"@"+ver)
                fi, err := os.Stat(entry)
                if err != nil || !fi.IsDir() {
                    ok = false
                    logger.Error("cache entry missing", map[string]interface{}{"pkg": pkg, "version": ver, "path": entry})
                    continue
                }
                d2, err := ammod.CommitDigestForCLI(entry, ver)
                if err != nil {
                    ok = false
                    logger.Error("digest compute failed", map[string]interface{}{"pkg": pkg, "version": ver, "error": err.Error()})
                    continue
                }
                if d2 != digest {
                    ok = false
                    logger.Error("digest mismatch", map[string]interface{}{"pkg": pkg, "version": ver})
                }
            }
        }
        if ok {
            logger.Info("ami.sum verified", nil)
        } else {
            // Fail with integrity violation exit code 3
            os.Stderr.WriteString("integrity violation: ami.sum does not match cache\n")
            os.Exit(ex.IntegrityViolationError)
        }
    },
}
