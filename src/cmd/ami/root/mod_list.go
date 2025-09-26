package root

import (
    "path/filepath"
    "sort"
    "strings"

    manifest "github.com/sam-caldwell/ami/src/ami/manifest"
    ammod "github.com/sam-caldwell/ami/src/ami/mod"
    "github.com/sam-caldwell/ami/src/internal/logger"
    "github.com/spf13/cobra"
)

func newModListCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "list",
        Short: "List cached packages",
        Example: `  ami mod list
  ami mod list --json`,
        Run: func(cmd *cobra.Command, args []string) {
            // Determine cache dir path (creating it if necessary)
            cacheDir, err := ammod.CacheDir()
            if err != nil {
                logger.Error(err.Error(), nil)
                return
            }
            // List entries in deterministic order
            items, err := ammod.ListCache(cacheDir)
            if err != nil {
                logger.Error(err.Error(), nil)
                return
            }
            sort.Strings(items)
            // Attempt to load ami.sum for digest lookup (best-effort)
            sum, _ := ammod.LoadSumForCLI("ami.sum")
            // Attempt to load ami.manifest for richer package info (best-effort)
            var manPkgs map[string]map[string]struct {
                Full   string
                Digest string
                Source string
            }
            if m, err := manifest.Load("ami.manifest"); err == nil {
                manPkgs = make(map[string]map[string]struct {
                    Full   string
                    Digest string
                    Source string
                })
                for _, p := range m.Packages {
                    base := filepath.Base(p.Name)
                    if _, ok := manPkgs[base]; !ok {
                        manPkgs[base] = make(map[string]struct {
                            Full   string
                            Digest string
                            Source string
                        })
                    }
                    manPkgs[base][p.Version] = struct {
                        Full   string
                        Digest string
                        Source string
                    }{Full: p.Name, Digest: p.Digest, Source: p.Source}
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
                    if digest != "" { data["digest"] = digest }
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
}

