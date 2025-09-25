package root

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"
    "github.com/sam-caldwell/ami/src/internal/logger"
)

var initForce bool

var cmdInit = &cobra.Command{
    Use:   "init",
    Short: "Initialize the AMI workspace",
    Run: func(cmd *cobra.Command, args []string) {
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
