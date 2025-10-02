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
func selectHighestSemver(tags []string, includePrerelease bool) (string, error) {
    type sv struct{ major, minor, patch int; pre string; raw string }
    parse := func(t string) (sv, bool) {
        s := strings.TrimPrefix(strings.TrimSpace(t), "v")
        parts := strings.SplitN(s, "-", 2)
        nums := strings.Split(parts[0], ".")
        if len(nums) != 3 { return sv{}, false }
        maj, err1 := atoi(nums[0])
        min, err2 := atoi(nums[1])
        pat, err3 := atoi(nums[2])
        if err1 != nil || err2 != nil || err3 != nil { return sv{}, false }
        pre := ""
        if len(parts) == 2 { pre = parts[1] }
        if !includePrerelease && pre != "" { return sv{}, false }
        return sv{maj, min, pat, pre, t}, true
    }
    var best *sv
    for _, t := range tags {
        v, ok := parse(t)
        if !ok { continue }
        if best == nil { best = &v; continue }
        if v.major != best.major {
            if v.major > best.major { *best = v }
            continue
        }
        if v.minor != best.minor {
            if v.minor > best.minor { *best = v }
            continue
        }
        if v.patch != best.patch {
            if v.patch > best.patch { *best = v }
            continue
        }
        // If majors, minors, patch equal, prefer no prerelease over prerelease
        if best.pre != "" && v.pre == "" { *best = v }
    }
    if best == nil { return "", fmt.Errorf("no semver tags") }
    return best.raw, nil
}

func atoi(s string) (int, error) { return strconv.Atoi(s) }
