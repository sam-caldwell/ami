package root

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/sam-caldwell/ami/src/internal/logger"
	"github.com/spf13/cobra"
	"runtime"
)

var initForce bool

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the AMI workspace",
		Example: `  # Initialize in an existing Git repo
  ami init

  # Initialize and create a Git repo if missing
  ami init --force

  # Emit JSON output
  ami --json init --force`,
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
			// Project name defaults to current directory name
			wd, _ := os.Getwd()
			projName := filepath.Base(wd)
			// include current machine os/arch in toolchain.compiler.env
			cur := runtime.GOOS + "/" + runtime.GOARCH
			content := "version: 1.0.0\n" +
				"project:\n" +
				"  name: " + projName + "\n" +
				"  version: 0.0.1\n" +
				"toolchain:\n" +
				"  compiler:\n" +
				"    concurrency: NUM_CPU\n" +
				"    target: ./build\n" +
				"    env:\n" +
				"      - { os: \"" + cur + "\" }\n" +
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

			// Ensure .gitignore contains ./build
			giPath := filepath.Join(".gitignore")
			want := "./build"
			if b, err := os.ReadFile(giPath); err == nil {
				s := string(b)
				has := false
				for _, line := range strings.Split(s, "\n") {
					if strings.TrimSpace(line) == want {
						has = true
						break
					}
				}
				if !has {
					f, e := os.OpenFile(giPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
					if e == nil {
						// ensure newline before appending if needed
						if len(s) > 0 && s[len(s)-1] != '\n' {
							_, _ = f.WriteString("\n")
						}
						_, _ = f.WriteString(want + "\n")
						_ = f.Close()
					}
				}
			} else {
				// create new .gitignore with ./build
				_ = os.WriteFile(giPath, []byte(want+"\n"), 0644)
			}
		},
	}
	cmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing files if present")
	return cmd
}
