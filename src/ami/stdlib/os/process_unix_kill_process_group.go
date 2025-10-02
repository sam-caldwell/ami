//go:build !windows

package os

import (
    goos "os"
    "syscall"
)

func killProcessGroup(p *goos.Process) error {
    // negative pid => kill process group
    return syscall.Kill(-p.Pid, syscall.SIGKILL)
}

