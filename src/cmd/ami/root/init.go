package root

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"

    git "github.com/go-git/go-git/v5"
    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/internal/logger"
)

var initForce bool

var cmdInit = &cobra.Command{
    Use:   "init",
    Short: "Initialize the AMI workspace",
    Run: func(cmd *cobra.Command, args []string) {
        // Require current directory to be a git repository unless --force is used.
        // If not a repo and --force, initialize a new repo.
        if _, err := os.Stat(filepath.Join(".git")); os.IsNotExist(err) {
            if !initForce {
                logger.Error("not a git repository (run `git init` or use --force)", nil)
                return
            }
            if _, ierr := git.PlainInit(".", false); ierr != nil {
                logger.Warn("git init failed", map[string]interface{}{"error": ierr.Error()})
            } else {
                logger.Info("initialized git repository", nil)
            }
        }
        wsPath := "ami.workspace"
        // Write ami.workspace
        if _, err := os.Stat(wsPath); err == nil && !initForce {
            logger.Error("ami.workspace already exists (use --force to overwrite)", nil)
            return
        }
        content := "version: 1.0.0\n" +
            "toolchain:\n" +
            "  compiler:\n" +
            "    concurrency: NUM_CPU\n" +
            "    target: ./build\n" +
            "    env: []\n" +
            "  linker: {}\n" +
            "  linter: {}\n" +
            "packages:\n" +
            "  - main:\n" +
            "      version: 0.0.1\n" +
            "      root: ./src\n" +
            "      import: []\n"
        if err := os.WriteFile(wsPath, []byte(content), 0644); err != nil {
            logger.Error(fmt.Sprintf("failed writing %s: %v", wsPath, err), nil)
            return
        }
        // Write src/main.ami
        if err := os.MkdirAll("src", 0755); err != nil {
            logger.Error(fmt.Sprintf("mkdir src: %v", err), nil)
            return
        }
        mainPath := filepath.Join("src", "main.ami")
        if _, err := os.Stat(mainPath); errors.Is(err, os.ErrNotExist) || initForce {
            mainSrc := "// AMI main entrypoint (scaffold)\n" +
                "// TODO: define your pipeline here.\n"
        if err := os.WriteFile(mainPath, []byte(mainSrc), 0644); err != nil {
            logger.Error(fmt.Sprintf("failed writing %s: %v", mainPath, err), nil)
            return
        }
        }
        logger.Info("workspace initialized", map[string]interface{}{"workspace": wsPath, "source": mainPath})
    },
}

func init() {
    cmdInit.Flags().BoolVar(&initForce, "force", false, "overwrite existing files if present")
}
