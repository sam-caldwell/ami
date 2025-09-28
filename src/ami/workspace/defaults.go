package workspace

import "runtime"

// DefaultWorkspace constructs a Workspace with initial defaults aligned to SPEC.
func DefaultWorkspace() Workspace {
    env := runtime.GOOS + "/" + runtime.GOARCH
    return Workspace{
        Version: "1.0.0",
        Toolchain: Toolchain{
            Compiler: Compiler{
                Concurrency: "NUM_CPU",
                Target:      "./build",
                Env:         []string{env},
                Backend:     "llvm",
                Options:     []string{"verbose"},
            },
            Linker: Linker{
                Options: []string{"Optimize: 0"},
            },
            Linter: Linter{
                Options: []string{"strict"},
            },
        },
        Packages: PackageList{
            {Key: "main", Package: Package{
                Name:    "newProject",
                Version: "0.0.1",
                Root:    "./src",
                Import:  []string{},
            }},
        },
    }
}
