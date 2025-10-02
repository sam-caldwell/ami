package main

import (
    "encoding/json"
    "fmt"
    "io"
    "io/fs"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "strconv"

    "github.com/sam-caldwell/ami/src/ami/exit"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// runModGet fetches a module from a local path into the package cache
// and updates ami.sum in the workspace. git+ssh is planned for later milestones.
 

 

// modGetGit clones a git repository (git+ssh or file+git) at a specific tag
// into the package cache and updates ami.sum accordingly.
 

// listGitTags returns tag names from a repo URL or local path using `git ls-remote --tags` (for remotes)
// or `git -C <path> tag` (for local file path). The returned tags are plain names (e.g., v1.2.3).
 

// selectHighestSemver chooses the highest SemVer tag from a list.
// If includePrerelease is false, tags with a pre-release suffix are excluded.
 
