package root

import (
    ammod "github.com/sam-caldwell/ami/src/ami/mod"
    ex "github.com/sam-caldwell/ami/src/internal/exit"
    "github.com/sam-caldwell/ami/src/internal/logger"
    "github.com/spf13/cobra"
    "os"
    "path/filepath"
)

func newModVerifyCmd() *cobra.Command {
    return &cobra.Command{
        Use:     "verify",
        Short:   "Verify ami.sum against cache",
        Example: `  ami mod verify`,
        Run: func(cmd *cobra.Command, args []string) {
            // Simple verification: ensure each cached entry exists; if git repo, re-compute digest
            cacheDir, err := ammod.CacheDir()
            if err != nil { logger.Error(err.Error(), nil); return }
            sum, err := ammod.LoadSumForCLI("ami.sum")
            if err != nil { logger.Error(err.Error(), nil); return }
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
}

